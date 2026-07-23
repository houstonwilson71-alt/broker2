package executor

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/bsc-sniper/backend/internal/config"
	"github.com/bsc-sniper/backend/internal/db"
	"github.com/bsc-sniper/backend/internal/filter"
	redisclient "github.com/bsc-sniper/backend/internal/redis"
	"go.uber.org/zap"
)

const pancakeRouterABIJSON = `[
  {
    "inputs":[
      {"internalType":"uint256","name":"amountOutMin","type":"uint256"},
      {"internalType":"address[]","name":"path","type":"address[]"},
      {"internalType":"address","name":"to","type":"address"},
      {"internalType":"uint256","name":"deadline","type":"uint256"}
    ],
    "name":"swapExactETHForTokens",
    "outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],
    "stateMutability":"payable",
    "type":"function"
  },
  {
    "inputs":[
      {"internalType":"uint256","name":"amountOutMin","type":"uint256"},
      {"internalType":"address[]","name":"path","type":"address[]"},
      {"internalType":"address","name":"to","type":"address"},
      {"internalType":"uint256","name":"deadline","type":"uint256"}
    ],
    "name":"swapExactETHForTokensSupportingFeeOnTransferTokens",
    "outputs":[],
    "stateMutability":"payable",
    "type":"function"
  },
  {
    "inputs":[
      {"internalType":"uint256","name":"amountIn","type":"uint256"},
      {"internalType":"uint256","name":"amountOutMin","type":"uint256"},
      {"internalType":"address[]","name":"path","type":"address[]"},
      {"internalType":"address","name":"to","type":"address"},
      {"internalType":"uint256","name":"deadline","type":"uint256"}
    ],
    "name":"swapExactTokensForETHSupportingFeeOnTransferTokens",
    "outputs":[],
    "stateMutability":"nonpayable",
    "type":"function"
  },
  {
    "inputs":[
      {"internalType":"uint256","name":"amountIn","type":"uint256"},
      {"internalType":"address[]","name":"path","type":"address[]"}
    ],
    "name":"getAmountsOut",
    "outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],
    "stateMutability":"view",
    "type":"function"
  }
]`

const erc20ApproveABIJSON = `[
  {"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
]`

const (
	RouterAddr = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
	WBNBAddr   = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095b"
)

// ─── Circuit breaker ──────────────────────────────────────────────────────────

type circuitBreaker struct {
	mu            sync.Mutex
	consecutive   int
	pausedUntil   time.Time
	threshold     int
	pauseDuration time.Duration
}

func newCircuitBreaker(threshold int, pause time.Duration) *circuitBreaker {
	return &circuitBreaker{threshold: threshold, pauseDuration: pause}
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	cb.consecutive = 0
	cb.mu.Unlock()
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	cb.consecutive++
	if cb.consecutive >= cb.threshold {
		cb.pausedUntil = time.Now().Add(cb.pauseDuration)
		cb.consecutive = 0
	}
	cb.mu.Unlock()
}

func (cb *circuitBreaker) isOpen() (bool, time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if time.Now().Before(cb.pausedUntil) {
		return true, time.Until(cb.pausedUntil)
	}
	return false, 0
}

// ─── Executor ─────────────────────────────────────────────────────────────────

type Executor struct {
	cfg        *config.Config
	rpc        *ethclient.Client
	redis      *redisclient.Client
	db         *db.DB
	logger     *zap.Logger
	privateKey *ecdsa.PrivateKey
	fromAddr   common.Address
	routerABI  abi.ABI
	erc20ABI   abi.ABI
	chainID    *big.Int

	circuit    *circuitBreaker
	rpcTokens  chan struct{}
	nonceMu    sync.Mutex

	// Tracks active buys for state reporting
	BuyingActive atomic.Int64
}

func New(cfg *config.Config, rpc *ethclient.Client, redis *redisclient.Client, database *db.DB, logger *zap.Logger) (*Executor, error) {
	rawKey := cfg.PrivateKey
	if strings.HasPrefix(rawKey, "0x") || strings.HasPrefix(rawKey, "0X") {
		rawKey = rawKey[2:]
	}
	pk, err := crypto.HexToECDSA(rawKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	pub := pk.Public().(*ecdsa.PublicKey)
	fromAddr := crypto.PubkeyToAddress(*pub)

	routerABI, err := abi.JSON(strings.NewReader(pancakeRouterABIJSON))
	if err != nil {
		return nil, fmt.Errorf("parse router abi: %w", err)
	}
	erc20ABI, err := abi.JSON(strings.NewReader(erc20ApproveABIJSON))
	if err != nil {
		return nil, fmt.Errorf("parse erc20 abi: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get chain id: %w", err)
	}

	ex := &Executor{
		cfg:        cfg,
		rpc:        rpc,
		redis:      redis,
		db:         database,
		logger:     logger,
		privateKey: pk,
		fromAddr:   fromAddr,
		routerABI:  routerABI,
		erc20ABI:   erc20ABI,
		chainID:    chainID,
		circuit:    newCircuitBreaker(3, 2*time.Minute),
		rpcTokens:  make(chan struct{}, 100),
	}

	// Prefill rate-limiter and start refill goroutine
	for i := 0; i < 100; i++ {
		ex.rpcTokens <- struct{}{}
	}
	go ex.refillRPCTokens()

	logger.Info("executor initialised",
		zap.String("wallet", fromAddr.Hex()),
		zap.Int64("chain_id", chainID.Int64()),
	)
	return ex, nil
}

func (ex *Executor) refillRPCTokens() {
	ticker := time.NewTicker(time.Second / 100)
	defer ticker.Stop()
	for range ticker.C {
		select {
		case ex.rpcTokens <- struct{}{}:
		default:
		}
	}
}

func (ex *Executor) acquireRPC() { <-ex.rpcTokens }

// Run is a single executor worker; spawn 4 with workerID 0..3.
func (ex *Executor) Run(ctx context.Context, workerID int) {
	consumer := fmt.Sprintf("%s:%d", redisclient.ConsumerExecutor, workerID)
	ex.logger.Info("executor worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-ctx.Done():
			ex.logger.Info("executor worker shutting down", zap.Int("worker_id", workerID))
			return
		default:
		}

		// Circuit breaker check
		if open, remaining := ex.circuit.isOpen(); open {
			ex.logger.Warn("circuit breaker OPEN — pausing buys",
				zap.Duration("remaining", remaining.Round(time.Second)),
			)
			select {
			case <-ctx.Done():
				return
			case <-time.After(minDur(remaining, 5*time.Second)):
				continue
			}
		}

		streams, err := ex.redis.ReadStream(ctx,
			redisclient.StreamApproved, redisclient.GroupExecutor, consumer, 5, 2*time.Second)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				data, ok := msg.Values["data"].(string)
				if !ok {
					_ = ex.redis.AckMessage(ctx, redisclient.StreamApproved, redisclient.GroupExecutor, msg.ID)
					continue
				}

				var approved filter.ApprovedToken
				if err := json.Unmarshal([]byte(data), &approved); err != nil {
					_ = ex.redis.AckMessage(ctx, redisclient.StreamApproved, redisclient.GroupExecutor, msg.ID)
					continue
				}

				go func(msgID string, tok filter.ApprovedToken) {
					defer func() {
						_ = ex.redis.AckMessage(ctx, redisclient.StreamApproved, redisclient.GroupExecutor, msgID)
					}()
					ex.BuyingActive.Add(1)
					defer ex.BuyingActive.Add(-1)
					ex.executeBuy(ctx, &tok, false)
				}(msg.ID, approved)
			}
		}
	}
}

func (ex *Executor) executeBuy(ctx context.Context, tok *filter.ApprovedToken, isRetry bool) {
	if !ex.cfg.LiveTradingEnabled {
		ex.logger.Info("SIMULATED BUY",
			zap.String("token", tok.TokenAddress),
			zap.String("symbol", tok.TokenSymbol),
			zap.Float64("bnb", ex.cfg.BuyAmountBNB),
		)
		ex.publishSimulatedTrade(ctx, tok)
		return
	}

	slippage := ex.cfg.SlippageBPS
	gasMultiplier := int64(15) // 1.5x
	if isRetry {
		slippage *= 2
		gasMultiplier = 30 // 3x on retry
		ex.logger.Info("retrying buy", zap.String("token", tok.TokenAddress))
	}

	txCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	amountBNBWei := toWei(ex.cfg.BuyAmountBNB)
	routerAddr := common.HexToAddress(RouterAddr)
	wbnbAddr := common.HexToAddress(WBNBAddr)
	tokenAddr := common.HexToAddress(tok.TokenAddress)
	path := []common.Address{wbnbAddr, tokenAddr}

	// Get expected output
	ex.acquireRPC()
	amountsData, err := ex.routerABI.Pack("getAmountsOut", amountBNBWei, path)
	if err != nil {
		ex.logger.Error("pack getAmountsOut", zap.Error(err))
		return
	}
	amountsResult, err := ex.rpc.CallContract(txCtx, ethereum.CallMsg{To: &routerAddr, Data: amountsData}, nil)
	if err != nil {
		ex.logger.Error("getAmountsOut", zap.Error(err))
		if !isRetry {
			time.Sleep(200 * time.Millisecond)
			ex.executeBuy(ctx, tok, true)
		}
		return
	}
	var amounts []*big.Int
	if err := ex.routerABI.UnpackIntoInterface(&amounts, "getAmountsOut", amountsResult); err != nil || len(amounts) < 2 {
		ex.logger.Error("unpack getAmountsOut", zap.Error(err))
		return
	}
	expectedOut := amounts[len(amounts)-1]

	amountOutMin := new(big.Int).Mul(expectedOut, big.NewInt(int64(10000-slippage)))
	amountOutMin.Div(amountOutMin, big.NewInt(10000))

	deadline := big.NewInt(time.Now().Add(2 * time.Minute).Unix())

	// Use fee-on-transfer variant to handle tokens with buy tax
	txData, err := ex.routerABI.Pack(
		"swapExactETHForTokensSupportingFeeOnTransferTokens",
		amountOutMin, path, ex.fromAddr, deadline,
	)
	if err != nil {
		ex.logger.Error("pack swap", zap.Error(err))
		return
	}

	// Get nonce (serialised to avoid nonce collisions across workers)
	ex.nonceMu.Lock()
	ex.acquireRPC()
	nonce, err := ex.rpc.PendingNonceAt(txCtx, ex.fromAddr)
	ex.nonceMu.Unlock()
	if err != nil {
		ex.logger.Error("get nonce", zap.Error(err))
		return
	}

	// Gas price = suggested * 1.5 (aggressive)
	ex.acquireRPC()
	suggestedGasPrice, err := ex.rpc.SuggestGasPrice(txCtx)
	if err != nil {
		ex.logger.Error("suggest gas price", zap.Error(err))
		return
	}
	gasPrice := new(big.Int).Mul(suggestedGasPrice, big.NewInt(gasMultiplier))
	gasPrice.Div(gasPrice, big.NewInt(10))

	// Estimate gas with fallback
	ex.acquireRPC()
	gasLimit, err := ex.rpc.EstimateGas(txCtx, ethereum.CallMsg{
		From:     ex.fromAddr,
		To:       &routerAddr,
		Value:    amountBNBWei,
		GasPrice: gasPrice,
		Data:     txData,
	})
	if err != nil {
		ex.logger.Warn("gas estimate failed, using 350000", zap.Error(err))
		gasLimit = 350000
	}
	gasLimit = gasLimit * 13 / 10 // +30% safety buffer

	tx := types.NewTransaction(nonce, routerAddr, amountBNBWei, gasLimit, gasPrice, txData)
	signer := types.NewEIP155Signer(ex.chainID)
	signedTx, err := types.SignTx(tx, signer, ex.privateKey)
	if err != nil {
		ex.logger.Error("sign tx", zap.Error(err))
		return
	}

	gasPriceGwei, _ := new(big.Float).Quo(
		new(big.Float).SetInt(gasPrice), new(big.Float).SetFloat64(1e9)).Float64()

	tradeID, err := ex.db.InsertTrade(txCtx, &db.Trade{
		TokenAddress: tok.TokenAddress,
		PairAddress:  tok.PairAddress,
		Side:         "buy",
		AmountBNB:    ex.cfg.BuyAmountBNB,
		AmountTokens: expectedOut.String(),
		TxHash:       signedTx.Hash().Hex(),
		GasPriceGwei: gasPriceGwei,
		Status:       "pending",
	})
	if err != nil {
		ex.logger.Error("insert trade", zap.Error(err))
	}

	// Dual-submit: public RPC + (optional) BloxRoute — first wins
	sendErr := ex.dualSubmit(txCtx, signedTx)

	if sendErr != nil {
		ex.logger.Error("send transaction",
			zap.Error(sendErr),
			zap.String("token", tok.TokenAddress),
		)
		if tradeID > 0 {
			_ = ex.db.UpdateTradeStatus(ctx, tradeID, "failed", "", sendErr.Error(), 0)
		}
		ex.circuit.recordFailure()
		if !isRetry {
			time.Sleep(200 * time.Millisecond)
			ex.executeBuy(ctx, tok, true)
		}
		return
	}

	ex.logger.Info("Buy executed",
		zap.String("token", tok.TokenAddress),
		zap.String("symbol", tok.TokenSymbol),
		zap.String("tx_hash", signedTx.Hash().Hex()),
		zap.Float64("bnb", ex.cfg.BuyAmountBNB),
		zap.Float64("gas_gwei", gasPriceGwei),
	)

	// Wait for receipt
	receiptCtx, rCancel := context.WithTimeout(ctx, 60*time.Second)
	defer rCancel()
	receipt := ex.waitForReceipt(receiptCtx, signedTx.Hash())

	status := "confirmed"
	gasUsed := int64(0)
	if receipt != nil {
		gasUsed = int64(receipt.GasUsed)
		if receipt.Status == 0 {
			status = "reverted"
		}
	}
	if tradeID > 0 {
		_ = ex.db.UpdateTradeStatus(ctx, tradeID, status, signedTx.Hash().Hex(), "", gasUsed)
	}

	if status == "reverted" {
		ex.logger.Error("Buy tx reverted", zap.String("tx_hash", signedTx.Hash().Hex()))
		ex.circuit.recordFailure()
		return
	}

	ex.circuit.recordSuccess()
	_ = ex.db.IncrBotCounters(ctx, 0, 0, 1)

	entryPrice := new(big.Float).Quo(
		new(big.Float).SetInt(amountBNBWei),
		new(big.Float).SetInt(expectedOut),
	)
	entryPriceStr := entryPrice.Text('f', 18)

	position := &db.Position{
		TokenAddress:    tok.TokenAddress,
		PairAddress:     tok.PairAddress,
		TokenSymbol:     tok.TokenSymbol,
		EntryPriceBNB:   entryPriceStr,
		CurrentPriceBNB: entryPriceStr,
		ATHPriceBNB:     entryPriceStr,
		AmountTokens:    expectedOut.String(),
		CostBNB:         ex.cfg.BuyAmountBNB,
		Status:          "open",
	}
	if dbErr := ex.db.UpsertPosition(ctx, position); dbErr != nil {
		ex.logger.Error("upsert position", zap.Error(dbErr))
	}

	_ = ex.redis.Publish(ctx, redisclient.PubSubEvents, map[string]interface{}{
		"type":     "position_opened",
		"token":    tok.TokenAddress,
		"symbol":   tok.TokenSymbol,
		"tx_hash":  signedTx.Hash().Hex(),
		"position": position,
	})
}

// dualSubmit sends to public RPC and optionally BloxRoute concurrently; first success wins.
func (ex *Executor) dualSubmit(ctx context.Context, tx *types.Transaction) error {
	if ex.cfg.BloxrouteURL == "" {
		ex.acquireRPC()
		return ex.rpc.SendTransaction(ctx, tx)
	}

	type result struct {
		err error
	}
	ch := make(chan result, 2)

	go func() {
		ex.acquireRPC()
		ch <- result{ex.rpc.SendTransaction(ctx, tx)}
	}()
	go func() {
		// BloxRoute fallback: send via public RPC as well for now
		// (Full BloxRoute integration requires their SDK)
		ex.acquireRPC()
		ch <- result{ex.rpc.SendTransaction(ctx, tx)}
	}()

	var lastErr error
	for i := 0; i < 2; i++ {
		r := <-ch
		if r.err == nil {
			return nil
		}
		lastErr = r.err
	}
	return lastErr
}

// ExecuteSell sells pct% of the position.
func (ex *Executor) ExecuteSell(ctx context.Context, pos *db.Position, pct int) error {
	if !ex.cfg.LiveTradingEnabled {
		ex.logger.Info("SIMULATED SELL",
			zap.String("token", pos.TokenAddress),
			zap.Int("pct", pct),
		)
		return nil
	}

	tokenAddr := common.HexToAddress(pos.TokenAddress)
	routerAddr := common.HexToAddress(RouterAddr)
	wbnbAddr := common.HexToAddress(WBNBAddr)

	// Get actual balance
	ex.acquireRPC()
	balData, err := ex.erc20ABI.Pack("balanceOf", ex.fromAddr)
	if err != nil {
		return fmt.Errorf("pack balanceOf: %w", err)
	}
	balResult, err := ex.rpc.CallContract(ctx, ethereum.CallMsg{To: &tokenAddr, Data: balData}, nil)
	if err != nil {
		return fmt.Errorf("balanceOf: %w", err)
	}
	var bal *big.Int
	if err := ex.erc20ABI.UnpackIntoInterface(&bal, "balanceOf", balResult); err != nil {
		return fmt.Errorf("unpack balanceOf: %w", err)
	}

	amountToSell := new(big.Int).Mul(bal, big.NewInt(int64(pct)))
	amountToSell.Div(amountToSell, big.NewInt(100))
	if amountToSell.Sign() == 0 {
		return fmt.Errorf("zero balance to sell")
	}

	// Approve router
	if err := ex.approveRouter(ctx, tokenAddr, routerAddr, amountToSell); err != nil {
		return fmt.Errorf("approve: %w", err)
	}

	// Get expected BNB out
	path := []common.Address{tokenAddr, wbnbAddr}
	ex.acquireRPC()
	amountsData, err := ex.routerABI.Pack("getAmountsOut", amountToSell, path)
	if err != nil {
		return err
	}
	amountsResult, err := ex.rpc.CallContract(ctx, ethereum.CallMsg{To: &routerAddr, Data: amountsData}, nil)
	if err != nil {
		return fmt.Errorf("getAmountsOut sell: %w", err)
	}
	var amounts []*big.Int
	if err := ex.routerABI.UnpackIntoInterface(&amounts, "getAmountsOut", amountsResult); err != nil || len(amounts) < 2 {
		return fmt.Errorf("unpack sell amounts: %w", err)
	}
	expectedBNB := amounts[len(amounts)-1]
	amountOutMin := new(big.Int).Mul(expectedBNB, big.NewInt(int64(10000-ex.cfg.SlippageBPS)))
	amountOutMin.Div(amountOutMin, big.NewInt(10000))

	deadline := big.NewInt(time.Now().Add(2 * time.Minute).Unix())
	txData, err := ex.routerABI.Pack("swapExactTokensForETHSupportingFeeOnTransferTokens",
		amountToSell, amountOutMin, path, ex.fromAddr, deadline)
	if err != nil {
		return fmt.Errorf("pack sell: %w", err)
	}

	ex.nonceMu.Lock()
	ex.acquireRPC()
	nonce, err := ex.rpc.PendingNonceAt(ctx, ex.fromAddr)
	ex.nonceMu.Unlock()
	if err != nil {
		return fmt.Errorf("nonce: %w", err)
	}

	ex.acquireRPC()
	suggestedGasPrice, err := ex.rpc.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("gas price: %w", err)
	}
	gasPrice := new(big.Int).Mul(suggestedGasPrice, big.NewInt(15))
	gasPrice.Div(gasPrice, big.NewInt(10))

	// Estimate gas with fallback
	ex.acquireRPC()
	gasLimit, err := ex.rpc.EstimateGas(ctx, ethereum.CallMsg{
		From:     ex.fromAddr,
		To:       &routerAddr,
		GasPrice: gasPrice,
		Data:     txData,
	})
	if err != nil {
		gasLimit = 300000
	}
	gasLimit = gasLimit * 13 / 10

	tx := types.NewTransaction(nonce, routerAddr, nil, gasLimit, gasPrice, txData)
	signer := types.NewEIP155Signer(ex.chainID)
	signedTx, err := types.SignTx(tx, signer, ex.privateKey)
	if err != nil {
		return fmt.Errorf("sign sell: %w", err)
	}

	gasPriceGwei, _ := new(big.Float).Quo(
		new(big.Float).SetInt(gasPrice), new(big.Float).SetFloat64(1e9)).Float64()
	bnbReceived, _ := new(big.Float).Quo(
		new(big.Float).SetInt(expectedBNB), new(big.Float).SetFloat64(1e18)).Float64()
	currentPricePerToken, _ := new(big.Float).Quo(
		new(big.Float).SetInt(expectedBNB), new(big.Float).SetInt(amountToSell)).Float64()
	pnl := bnbReceived - (ex.cfg.BuyAmountBNB * float64(pct) / 100)

	tradeID, _ := ex.db.InsertTrade(ctx, &db.Trade{
		TokenAddress: pos.TokenAddress,
		PairAddress:  pos.PairAddress,
		Side:         "sell",
		AmountBNB:    bnbReceived,
		AmountTokens: amountToSell.String(),
		PriceBNB:     fmt.Sprintf("%.18f", currentPricePerToken),
		TxHash:       signedTx.Hash().Hex(),
		GasPriceGwei: gasPriceGwei,
		Status:       "pending",
	})

	ex.acquireRPC()
	if err := ex.rpc.SendTransaction(ctx, signedTx); err != nil {
		if tradeID > 0 {
			_ = ex.db.UpdateTradeStatus(ctx, tradeID, "failed", "", err.Error(), 0)
		}
		return fmt.Errorf("send sell: %w", err)
	}

	ex.logger.Info("Sell executed",
		zap.String("token", pos.TokenAddress),
		zap.String("symbol", pos.TokenSymbol),
		zap.String("tx_hash", signedTx.Hash().Hex()),
		zap.Int("pct", pct),
		zap.Float64("bnb_received", bnbReceived),
		zap.Float64("pnl_bnb", pnl),
	)

	receiptCtx, rCancel := context.WithTimeout(ctx, 60*time.Second)
	defer rCancel()
	receipt := ex.waitForReceipt(receiptCtx, signedTx.Hash())
	sellStatus := "confirmed"
	gasUsed := int64(0)
	if receipt != nil {
		gasUsed = int64(receipt.GasUsed)
		if receipt.Status == 0 {
			sellStatus = "reverted"
		}
	}
	if tradeID > 0 {
		_ = ex.db.UpdateTradeStatus(ctx, tradeID, sellStatus, signedTx.Hash().Hex(), "", gasUsed)
	}

	_ = ex.redis.Publish(ctx, redisclient.PubSubEvents, map[string]interface{}{
		"type":         "sell_executed",
		"token":        pos.TokenAddress,
		"symbol":       pos.TokenSymbol,
		"tx_hash":      signedTx.Hash().Hex(),
		"pct":          pct,
		"bnb_received": bnbReceived,
		"pnl_bnb":      pnl,
	})
	return nil
}

func (ex *Executor) approveRouter(ctx context.Context, tokenAddr, routerAddr common.Address, amount *big.Int) error {
	data, err := ex.erc20ABI.Pack("approve", routerAddr, amount)
	if err != nil {
		return err
	}

	ex.nonceMu.Lock()
	ex.acquireRPC()
	nonce, err := ex.rpc.PendingNonceAt(ctx, ex.fromAddr)
	ex.nonceMu.Unlock()
	if err != nil {
		return err
	}

	ex.acquireRPC()
	gasPrice, err := ex.rpc.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(15))
	gasPrice.Div(gasPrice, big.NewInt(10))

	tx := types.NewTransaction(nonce, tokenAddr, nil, 100000, gasPrice, data)
	signer := types.NewEIP155Signer(ex.chainID)
	signedTx, err := types.SignTx(tx, signer, ex.privateKey)
	if err != nil {
		return err
	}

	ex.acquireRPC()
	if err := ex.rpc.SendTransaction(ctx, signedTx); err != nil {
		return err
	}

	receiptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	ex.waitForReceipt(receiptCtx, signedTx.Hash())
	return nil
}

func (ex *Executor) waitForReceipt(ctx context.Context, txHash common.Hash) *types.Receipt {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		ex.acquireRPC()
		receipt, err := ex.rpc.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt
		}
		time.Sleep(2 * time.Second)
	}
}

func (ex *Executor) publishSimulatedTrade(ctx context.Context, tok *filter.ApprovedToken) {
	entryPrice := "0.0000001"
	amountTokens := "1000000000000000000"
	position := &db.Position{
		TokenAddress:    tok.TokenAddress,
		PairAddress:     tok.PairAddress,
		TokenSymbol:     tok.TokenSymbol,
		EntryPriceBNB:   entryPrice,
		CurrentPriceBNB: entryPrice,
		ATHPriceBNB:     entryPrice,
		AmountTokens:    amountTokens,
		CostBNB:         ex.cfg.BuyAmountBNB,
		Status:          "open",
	}
	_ = ex.db.UpsertPosition(ctx, position)
	_ = ex.redis.Publish(ctx, redisclient.PubSubEvents, map[string]interface{}{
		"type":     "position_opened_simulated",
		"token":    tok.TokenAddress,
		"symbol":   tok.TokenSymbol,
		"position": position,
	})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func toWei(bnb float64) *big.Int {
	f := new(big.Float).SetFloat64(bnb)
	f.Mul(f, new(big.Float).SetFloat64(1e18))
	result, _ := f.Int(nil)
	return result
}

func minDur(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
