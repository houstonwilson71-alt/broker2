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
	WBNBAddr   = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
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

	circuit   *circuitBreaker
	rpcTokens chan struct{}
	nonceMu   sync.Mutex

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

// buildBuyPath returns the token path for a buy transaction.
// Single-hop if quoteToken == WBNB; two-hop otherwise.
func buildBuyPath(quoteToken string, tokenAddr common.Address) []common.Address {
	wbnb := common.HexToAddress(WBNBAddr)
	if strings.EqualFold(quoteToken, WBNBAddr) || quoteToken == "" {
		return []common.Address{wbnb, tokenAddr}
	}
	return []common.Address{wbnb, common.HexToAddress(quoteToken), tokenAddr}
}

// buildSellPath returns the token path for a sell transaction (reverse of buy path).
func buildSellPath(quoteToken string, tokenAddr common.Address) []common.Address {
	wbnb := common.HexToAddress(WBNBAddr)
	if strings.EqualFold(quoteToken, WBNBAddr) || quoteToken == "" {
		return []common.Address{tokenAddr, wbnb}
	}
	return []common.Address{tokenAddr, common.HexToAddress(quoteToken), wbnb}
}

func (ex *Executor) executeBuy(ctx context.Context, tok *filter.ApprovedToken, isRetry bool) {
	if !ex.cfg.LiveTradingEnabled {
		ex.logger.Info("SIMULATED BUY",
			zap.String("token", tok.TokenAddress),
			zap.String("symbol", tok.TokenSymbol),
			zap.String("quote", tok.QuoteSymbol),
			zap.Float64("bnb", ex.cfg.BuyAmountBNB),
		)
		ex.publishSimulatedTrade(ctx, tok)
		return
	}

	// Gas reserve: wallet must keep at least 0.01 BNB for sells / gas.
	const minGasReserveBNB = 0.01
	balCtx, bCancel := context.WithTimeout(ctx, 5*time.Second)
	defer bCancel()
	ex.acquireRPC()
	bal, err := ex.rpc.BalanceAt(balCtx, ex.fromAddr, nil)
	if err == nil {
		balF, _ := new(big.Float).Quo(new(big.Float).SetInt(bal), new(big.Float).SetFloat64(1e18)).Float64()
		ex.logger.Info("wallet balance before buy",
			zap.String("token", tok.TokenAddress),
			zap.String("symbol", tok.TokenSymbol),
			zap.Float64("bnb_balance", balF),
			zap.Float64("buy_bnb", ex.cfg.BuyAmountBNB),
		)
		if balF < minGasReserveBNB {
			ex.logger.Warn("WALLET ARMOR: BNB balance below gas reserve, skipping buy",
				zap.String("token", tok.TokenAddress),
				zap.String("symbol", tok.TokenSymbol),
				zap.Float64("bnb_balance", balF),
				zap.Float64("min_reserve", minGasReserveBNB),
			)
			_, _ = ex.db.InsertTrade(ctx, &db.Trade{
				TokenAddress: tok.TokenAddress,
				PairAddress:  tok.PairAddress,
				Side:         "buy",
				AmountBNB:    0,
				Status:       "failed",
				ErrorMsg:     fmt.Sprintf("gas_reserve: balance %.6f BNB < %.4f", balF, minGasReserveBNB),
			})
			return
		}
	}

	slippage := ex.cfg.SlippageBPS
	gasMultiplier := int64(15) // 1.5x default
	if isRetry {
		gasMultiplier = 20 // 2.0x retry
		ex.logger.Info("retrying buy with 2.0x gas", zap.String("token", tok.TokenAddress))
	}

	txCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	amountBNBWei := toWei(ex.cfg.BuyAmountBNB)
	routerAddr := common.HexToAddress(RouterAddr)
	tokenAddr := common.HexToAddress(tok.TokenAddress)

	// Build path: single-hop for WBNB, two-hop for stablecoins/CAKE/ETH
	path := buildBuyPath(tok.QuoteToken, tokenAddr)

	ex.logger.Info("buy path",
		zap.String("token", tok.TokenSymbol),
		zap.String("quote", tok.QuoteSymbol),
		zap.Int("hops", len(path)-1),
	)

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

	txData, err := ex.routerABI.Pack(
		"swapExactETHForTokensSupportingFeeOnTransferTokens",
		amountOutMin, path, ex.fromAddr, deadline,
	)
	if err != nil {
		ex.logger.Error("pack swap", zap.Error(err))
		return
	}

	ex.nonceMu.Lock()
	ex.acquireRPC()
	nonce, err := ex.rpc.PendingNonceAt(txCtx, ex.fromAddr)
	ex.nonceMu.Unlock()
	if err != nil {
		ex.logger.Error("get nonce", zap.Error(err))
		return
	}

	ex.acquireRPC()
	suggestedGasPrice, err := ex.rpc.SuggestGasPrice(txCtx)
	if err != nil {
		ex.logger.Error("suggest gas price", zap.Error(err))
		return
	}
	gasPrice := new(big.Int).Mul(suggestedGasPrice, big.NewInt(gasMultiplier))
	gasPrice.Div(gasPrice, big.NewInt(10))

	// Estimate gas — multi-hop needs higher limit
	defaultGasLimit := uint64(350000)
	if len(path) > 2 {
		defaultGasLimit = 500000 // two-hop swap
	}

	ex.acquireRPC()
	gasLimit, err := ex.rpc.EstimateGas(txCtx, ethereum.CallMsg{
		From:     ex.fromAddr,
		To:       &routerAddr,
		Value:    amountBNBWei,
		GasPrice: gasPrice,
		Data:     txData,
	})
	if err != nil {
		ex.logger.Warn("gas estimate failed, using default",
			zap.Uint64("default", defaultGasLimit),
			zap.Error(err),
		)
		gasLimit = defaultGasLimit
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
		PriceBNB:     "0", // updated post-confirm; "0" avoids NUMERIC cast error
		TxHash:       signedTx.Hash().Hex(),
		GasPriceGwei: gasPriceGwei,
		Status:       "pending",
	})
	if err != nil {
		ex.logger.Error("insert trade", zap.Error(err))
	}

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
			ex.executeBuy(ctx, tok, true)
		}
		return
	}

	ex.logger.Info("Buy executed",
		zap.String("token", tok.TokenAddress),
		zap.String("symbol", tok.TokenSymbol),
		zap.String("quote", tok.QuoteSymbol),
		zap.String("tx_hash", signedTx.Hash().Hex()),
		zap.Float64("bnb", ex.cfg.BuyAmountBNB),
		zap.Float64("gas_gwei", gasPriceGwei),
		zap.Int("path_hops", len(path)-1),
	)

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
		QuoteToken:      tok.QuoteToken,
		EntryPriceBNB:   entryPriceStr,
		CurrentPriceBNB: entryPriceStr,
		ATHPriceBNB:     entryPriceStr,
		AmountTokens:    expectedOut.String(),
		CostBNB:         ex.cfg.BuyAmountBNB,
		Status:          "bought",
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

	// Post-buy verification: immediately simulate selling 100% of received tokens.
	// If the sell would revert or return <90% of the BNB spent, try an emergency sell
	// with 50% slippage and 2.0x gas. If that fails, mark the position as unsellable.
	ex.postBuyVerification(ctx, position, signedTx.Hash().Hex())
}

// postBuyVerification simulates a full sell of the just-bought position and,
// if the round-trip is bad, attempts an emergency dump.
func (ex *Executor) postBuyVerification(ctx context.Context, pos *db.Position, buyTxHash string) {
	tokenAddr := common.HexToAddress(pos.TokenAddress)
	routerAddr := common.HexToAddress(RouterAddr)
	sellPath := buildSellPath(pos.QuoteToken, tokenAddr)

	// Read actual token balance received.
	balData, err := ex.erc20ABI.Pack("balanceOf", ex.fromAddr)
	if err != nil {
		ex.logger.Error("postbuy pack balanceOf", zap.Error(err), zap.String("token", pos.TokenAddress))
		return
	}
	ex.acquireRPC()
	balResult, err := ex.rpc.CallContract(ctx, ethereum.CallMsg{To: &tokenAddr, Data: balData}, nil)
	if err != nil {
		ex.logger.Error("postbuy balanceOf", zap.Error(err), zap.String("token", pos.TokenAddress))
		return
	}
	var bal *big.Int
	if err := ex.erc20ABI.UnpackIntoInterface(&bal, "balanceOf", balResult); err != nil {
		ex.logger.Error("postbuy unpack balanceOf", zap.Error(err), zap.String("token", pos.TokenAddress))
		return
	}
	if bal == nil || bal.Sign() == 0 {
		ex.logger.Warn("postbuy zero balance", zap.String("token", pos.TokenAddress))
		return
	}

	amountsData, err := ex.routerABI.Pack("getAmountsOut", bal, sellPath)
	if err != nil {
		ex.logger.Error("postbuy pack getAmountsOut", zap.Error(err), zap.String("token", pos.TokenAddress))
		return
	}
	ex.acquireRPC()
	amountsResult, err := ex.rpc.CallContract(ctx, ethereum.CallMsg{To: &routerAddr, Data: amountsData}, nil)
	if err != nil {
		ex.logger.Warn("postbuy sell simulation reverted",
			zap.Error(err),
			zap.String("token", pos.TokenAddress),
			zap.String("symbol", pos.TokenSymbol),
		)
		ex.tryEmergencyDump(ctx, pos, bal, buyTxHash)
		return
	}
	var amounts []*big.Int
	if err := ex.routerABI.UnpackIntoInterface(&amounts, "getAmountsOut", amountsResult); err != nil || len(amounts) < 2 {
		ex.logger.Warn("postbuy sell simulation unpack failed",
			zap.Error(err),
			zap.String("token", pos.TokenAddress),
		)
		ex.tryEmergencyDump(ctx, pos, bal, buyTxHash)
		return
	}
	bnbBack := amounts[len(amounts)-1]
	bnbIn := toWei(ex.cfg.BuyAmountBNB)
	efficiency, _ := new(big.Float).Quo(
		new(big.Float).SetInt(bnbBack),
		new(big.Float).SetInt(bnbIn),
	).Float64()

	ex.logger.Info("postbuy sell simulation",
		zap.String("token", pos.TokenAddress),
		zap.String("symbol", pos.TokenSymbol),
		zap.Float64("efficiency", efficiency),
	)

	if efficiency < 0.90 {
		ex.logger.Warn("postbuy efficiency below 0.90 — emergency dumping",
			zap.String("token", pos.TokenAddress),
			zap.String("symbol", pos.TokenSymbol),
			zap.Float64("efficiency", efficiency),
		)
		ex.tryEmergencyDump(ctx, pos, bal, buyTxHash)
	}
}

// tryEmergencyDump attempts to sell 100% of the position with 50% slippage and 2.0x gas.
// If the dump fails, the position is marked as unsellable.
func (ex *Executor) tryEmergencyDump(ctx context.Context, pos *db.Position, amount *big.Int, buyTxHash string) {
	if err := ex.ExecuteSellCustom(ctx, pos, 100, "unsellable_dump", 5000, 20); err != nil {
		ex.logger.Error("emergency dump failed — marking unsellable",
			zap.Error(err),
			zap.String("token", pos.TokenAddress),
			zap.String("symbol", pos.TokenSymbol),
		)
		pos.Status = "unsellable"
		if dbErr := ex.db.UpsertPosition(ctx, pos); dbErr != nil {
			ex.logger.Error("mark unsellable failed", zap.Error(dbErr), zap.String("token", pos.TokenAddress))
		}
		return
	}
	ex.logger.Info("emergency dump succeeded",
		zap.String("token", pos.TokenAddress),
		zap.String("symbol", pos.TokenSymbol),
	)
}

// dualSubmit sends to public RPC and optionally BloxRoute concurrently.
func (ex *Executor) dualSubmit(ctx context.Context, tx *types.Transaction) error {
	if ex.cfg.BloxrouteURL == "" {
		ex.acquireRPC()
		return ex.rpc.SendTransaction(ctx, tx)
	}

	type result struct{ err error }
	ch := make(chan result, 2)

	go func() {
		ex.acquireRPC()
		ch <- result{ex.rpc.SendTransaction(ctx, tx)}
	}()
	go func() {
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

// ExecuteSell sells pct% of the position using the correct reverse path with
// the standard 5% slippage and 1.5x gas (plus extra boost for TP exits).
// sellType is "tp_50", "tp_300", "trailing_sl", "breakeven_sl", "breakeven_sl_partial", or "force".
func (ex *Executor) ExecuteSell(ctx context.Context, pos *db.Position, pct int, sellType string) error {
	gasMultiplier := int64(15)
	if strings.HasPrefix(sellType, "tp_") || sellType == "trailing_sl" {
		gasMultiplier = 225 // 1.5x * 1.5x = 2.25x for TP sells
	}
	return ex.ExecuteSellCustom(ctx, pos, pct, sellType, 500, gasMultiplier)
}

// ExecuteSellCustom sells pct% of the position with configurable slippage (basis
// points) and gas multiplier (times 10; e.g. 15 = 1.5x). Used by the standard monitor
// exits and by the post-buy emergency dump.
func (ex *Executor) ExecuteSellCustom(ctx context.Context, pos *db.Position, pct int, sellType string, slippageBps int, gasMultiplier int64) error {
	if !ex.cfg.LiveTradingEnabled {
		ex.logger.Info("SIMULATED SELL",
			zap.String("token", pos.TokenAddress),
			zap.Int("pct", pct),
		)
		return nil
	}

	tokenAddr := common.HexToAddress(pos.TokenAddress)
	routerAddr := common.HexToAddress(RouterAddr)

	// Determine sell path (stored in position or default to WBNB)
	// For simplicity, we read the pair's tokens from the position.
	// The sell path mirrors the buy path.
	// pos.PairAddress is used only for identification; we rely on the token router.
	sellPath := buildSellPath(pos.QuoteToken, tokenAddr)

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

	if err := ex.approveRouter(ctx, tokenAddr, routerAddr, amountToSell); err != nil {
		return fmt.Errorf("approve: %w", err)
	}

	// Get expected BNB out
	ex.acquireRPC()
	amountsData, err := ex.routerABI.Pack("getAmountsOut", amountToSell, sellPath)
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
	// Slippage protection: amountOutMin = expected * (10000 - slippageBps) / 10000
	amountOutMin := new(big.Int).Mul(expectedBNB, big.NewInt(int64(10000-slippageBps)))
	amountOutMin.Div(amountOutMin, big.NewInt(10000))

	deadline := big.NewInt(time.Now().Add(2 * time.Minute).Unix())
	txData, err := ex.routerABI.Pack("swapExactTokensForETHSupportingFeeOnTransferTokens",
		amountToSell, amountOutMin, sellPath, ex.fromAddr, deadline)
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
	gasPrice := new(big.Int).Mul(suggestedGasPrice, big.NewInt(gasMultiplier))
	gasPrice.Div(gasPrice, big.NewInt(10))

	defaultGasLimit := uint64(300000)
	if len(sellPath) > 2 {
		defaultGasLimit = 500000
	}
	ex.acquireRPC()
	gasLimit, err := ex.rpc.EstimateGas(ctx, ethereum.CallMsg{
		From:     ex.fromAddr,
		To:       &routerAddr,
		GasPrice: gasPrice,
		Data:     txData,
	})
	if err != nil {
		gasLimit = defaultGasLimit
	}
	gasLimit = gasLimit * 13 / 10

	bnbReceived, _ := new(big.Float).Quo(
		new(big.Float).SetInt(expectedBNB), new(big.Float).SetFloat64(1e18)).Float64()
	currentPricePerToken, _ := new(big.Float).Quo(
		new(big.Float).SetInt(expectedBNB), new(big.Float).SetInt(amountToSell)).Float64()
	pnl := bnbReceived - (ex.cfg.BuyAmountBNB * float64(pct) / 100)

	// Attempt the sell once; if it fails or reverts, retry with 1.5x gas.
	var signedTx *types.Transaction
	signer := types.NewEIP155Signer(ex.chainID)
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(15))
			gasPrice.Div(gasPrice, big.NewInt(10))
			ex.nonceMu.Lock()
			ex.acquireRPC()
			nonce, _ = ex.rpc.PendingNonceAt(ctx, ex.fromAddr)
			ex.nonceMu.Unlock()
			ex.logger.Warn("sell retry with 1.5x gas",
				zap.String("token", pos.TokenAddress),
				zap.String("symbol", pos.TokenSymbol),
				zap.Int("attempt", attempt+1),
			)
		}

		tx := types.NewTransaction(nonce, routerAddr, nil, gasLimit, gasPrice, txData)
		var signErr error
		signedTx, signErr = types.SignTx(tx, signer, ex.privateKey)
		if signErr != nil {
			if attempt == 0 {
				return fmt.Errorf("sign sell: %w", signErr)
			}
			break
		}

		ex.acquireRPC()
		sendErr := ex.rpc.SendTransaction(ctx, signedTx)
		if sendErr == nil {
			receiptCtx, rCancel := context.WithTimeout(ctx, 60*time.Second)
			receipt := ex.waitForReceipt(receiptCtx, signedTx.Hash())
			rCancel()
			if receipt != nil && receipt.Status == 1 {
				break // success
			}
			if receipt != nil && receipt.Status == 0 {
				ex.logger.Warn("sell tx reverted, retrying with 1.5x gas",
					zap.String("token", pos.TokenAddress),
					zap.String("tx_hash", signedTx.Hash().Hex()),
				)
			}
		} else {
			ex.logger.Warn("send sell failed, retrying with 1.5x gas",
				zap.String("token", pos.TokenAddress),
				zap.Error(sendErr),
			)
		}
		if attempt == 1 {
			if sendErr != nil {
				return fmt.Errorf("send sell failed after retry: %w", sendErr)
			}
			return fmt.Errorf("sell tx failed after retry (reverted or receipt not found)")
		}
	}

	if signedTx == nil {
		return fmt.Errorf("sell transaction could not be built")
	}

	gasPriceGwei, _ := new(big.Float).Quo(
		new(big.Float).SetInt(gasPrice), new(big.Float).SetFloat64(1e9)).Float64()

	ex.logger.Info("Sell executed",
		zap.String("token", pos.TokenAddress),
		zap.String("symbol", pos.TokenSymbol),
		zap.String("tx_hash", signedTx.Hash().Hex()),
		zap.Int("pct", pct),
		zap.String("sell_type", sellType),
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

	_, _ = ex.db.InsertTrade(ctx, &db.Trade{
		TokenAddress: pos.TokenAddress,
		PairAddress:  pos.PairAddress,
		Side:         "sell",
		AmountBNB:    bnbReceived,
		AmountTokens: amountToSell.String(),
		PriceBNB:     fmt.Sprintf("%.18f", currentPricePerToken),
		TxHash:       signedTx.Hash().Hex(),
		GasPriceGwei: gasPriceGwei,
		Status:       sellStatus,
		GasUsed:      gasUsed,
	})

	// Update position status so the DB reflects reality even for emergency/force sells.
	if sellStatus == "confirmed" {
		if pct >= 100 {
			pos.Status = "closed"
			now := time.Now()
			pos.ClosedAt = &now
		} else {
			pos.Status = "partial"
		}
		pos.RealizedPnlBNB += pnl
		if amtBig, ok2 := new(big.Int).SetString(pos.AmountTokens, 10); ok2 {
			remaining := new(big.Int).Sub(amtBig, amountToSell)
			if remaining.Sign() < 0 {
				remaining = big.NewInt(0)
			}
			pos.AmountTokens = remaining.String()
		}
		if dbErr := ex.db.UpsertPosition(ctx, pos); dbErr != nil {
			ex.logger.Error("update position after sell", zap.Error(dbErr), zap.String("token", pos.TokenAddress))
		}
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
		QuoteToken:      tok.QuoteToken,
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
