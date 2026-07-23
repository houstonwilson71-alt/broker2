package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/bsc-sniper/backend/internal/config"
	"github.com/bsc-sniper/backend/internal/db"
	"github.com/bsc-sniper/backend/internal/listener"
	redisclient "github.com/bsc-sniper/backend/internal/redis"
	"go.uber.org/zap"
)

const pairABIJSON = `[
  {"inputs":[],"name":"getReserves","outputs":[{"internalType":"uint112","name":"_reserve0","type":"uint112"},{"internalType":"uint112","name":"_reserve1","type":"uint112"},{"internalType":"uint32","name":"_blockTimestampLast","type":"uint32"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}
]`

const erc20ABIJSON = `[
  {"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
]`

const routerABIJSON = `[
  {"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"}
]`

const (
	RouterAddress = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
	WBNBAddr      = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095b"
)

// ─── BNB price cache (shared across all workers) ──────────────────────────────

var (
	cachedBNBPrice   float64
	cachedBNBPriceAt time.Time
	bnbPriceMu       sync.RWMutex
)

// ApprovedToken is the event published to stream:approved_tokens.
type ApprovedToken struct {
	TokenAddress string  `json:"token_address"`
	PairAddress  string  `json:"pair_address"`
	TokenSymbol  string  `json:"token_symbol"`
	TokenName    string  `json:"token_name"`
	Decimals     int     `json:"decimals"`
	LiquidityUSD float64 `json:"liquidity_usd"`
	Source       string  `json:"source"`
	BlockNumber  uint64  `json:"block_number"`
	Timestamp    int64   `json:"timestamp"`
}

// Engine processes pairs through safety filters.
type Engine struct {
	cfg    *config.Config
	rpc    *ethclient.Client
	redis  *redisclient.Client
	db     *db.DB
	logger *zap.Logger

	pairABI   abi.ABI
	erc20ABI  abi.ABI
	routerABI abi.ABI

	// rate limiter: max 100 RPC calls/second shared across all filter goroutines
	rpcTokens chan struct{}
	// active processing counter (for state reporting)
	Active atomic.Int64
}

func New(cfg *config.Config, rpc *ethclient.Client, redis *redisclient.Client, database *db.DB, logger *zap.Logger) (*Engine, error) {
	pairABI, err := abi.JSON(strings.NewReader(pairABIJSON))
	if err != nil {
		return nil, fmt.Errorf("parse pair abi: %w", err)
	}
	erc20ABI, err := abi.JSON(strings.NewReader(erc20ABIJSON))
	if err != nil {
		return nil, fmt.Errorf("parse erc20 abi: %w", err)
	}
	routerABI, err := abi.JSON(strings.NewReader(routerABIJSON))
	if err != nil {
		return nil, fmt.Errorf("parse router abi: %w", err)
	}

	e := &Engine{
		cfg:       cfg,
		rpc:       rpc,
		redis:     redis,
		db:        database,
		logger:    logger,
		pairABI:   pairABI,
		erc20ABI:  erc20ABI,
		routerABI: routerABI,
		rpcTokens: make(chan struct{}, 100),
	}

	// Prefill the rate-limiter bucket and start refill goroutine
	for i := 0; i < 100; i++ {
		e.rpcTokens <- struct{}{}
	}
	go e.refillRPCTokens()

	return e, nil
}

func (e *Engine) refillRPCTokens() {
	ticker := time.NewTicker(time.Second / 100) // 100 tokens per second
	defer ticker.Stop()
	for range ticker.C {
		select {
		case e.rpcTokens <- struct{}{}:
		default:
		}
	}
}

func (e *Engine) acquireRPC() {
	<-e.rpcTokens
}

// Run is a single filter worker; spawn 4 with workerID 0..3.
func (e *Engine) Run(ctx context.Context, workerID int) {
	consumer := fmt.Sprintf("%s:%d", redisclient.ConsumerFilter, workerID)
	e.logger.Info("filter worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("filter worker shutting down", zap.Int("worker_id", workerID))
			return
		default:
		}

		streams, err := e.redis.ReadStream(ctx,
			redisclient.StreamNewPairs, redisclient.GroupFilter, consumer, 5, 2*time.Second)
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
					_ = e.redis.AckMessage(ctx, redisclient.StreamNewPairs, redisclient.GroupFilter, msg.ID)
					continue
				}

				var event listener.NewPairEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					_ = e.redis.AckMessage(ctx, redisclient.StreamNewPairs, redisclient.GroupFilter, msg.ID)
					continue
				}

				_ = e.db.IncrBotCounters(ctx, 1, 0, 0)

				go func(msgID string, ev listener.NewPairEvent) {
					e.Active.Add(1)
					defer e.Active.Add(-1)
					e.processEvent(ctx, &ev)
					_ = e.redis.AckMessage(ctx, redisclient.StreamNewPairs, redisclient.GroupFilter, msgID)
				}(msg.ID, event)
			}
		}
	}
}

func (e *Engine) processEvent(ctx context.Context, event *listener.NewPairEvent) {
	// Strict 2.5 second total deadline
	filterCtx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()

	// Determine meme token
	wbnb := strings.ToLower(WBNBAddr)
	memeToken := event.MemeToken
	if memeToken == "" {
		// Fallback compute
		if strings.ToLower(event.Token0) == wbnb {
			memeToken = event.Token1
		} else {
			memeToken = event.Token0
		}
	}

	result := &db.FilterResult{
		TokenAddress: memeToken,
		PairAddress:  event.PairAddress,
	}

	type checkResult struct {
		name string
		err  error
		val  interface{}
	}

	liquidityCh := make(chan checkResult, 1)
	ageCh := make(chan checkResult, 1)
	tokenInfoCh := make(chan checkResult, 1)
	honeypotCh := make(chan checkResult, 1)

	// All checks run in parallel with shared 2.5s deadline
	go func() {
		checkCtx, c := context.WithTimeout(filterCtx, 2*time.Second)
		defer c()
		liq, err := e.getLiquidityUSD(checkCtx, event.PairAddress, memeToken)
		liquidityCh <- checkResult{"liquidity", err, liq}
	}()
	go func() {
		checkCtx, c := context.WithTimeout(filterCtx, 2*time.Second)
		defer c()
		age, err := e.getTokenAge(checkCtx, event.BlockNumber)
		ageCh <- checkResult{"age", err, age}
	}()
	go func() {
		checkCtx, c := context.WithTimeout(filterCtx, 2*time.Second)
		defer c()
		info, err := e.getTokenInfo(checkCtx, memeToken)
		tokenInfoCh <- checkResult{"tokeninfo", err, info}
	}()
	go func() {
		checkCtx, c := context.WithTimeout(filterCtx, 2*time.Second)
		defer c()
		hp, err := e.checkHoneypot(checkCtx, memeToken, event.PairAddress)
		honeypotCh <- checkResult{"honeypot", err, hp}
	}()

	failReasons := []string{}
	passed := true
	var tInfo tokenInfoVal

	// Collect all four results within deadline
	for i := 0; i < 4; i++ {
		var res checkResult
		select {
		case res = <-liquidityCh:
		case res = <-ageCh:
		case res = <-tokenInfoCh:
		case res = <-honeypotCh:
		case <-filterCtx.Done():
			e.logger.Warn("filter stage timeout",
				zap.String("pair", event.PairAddress),
				zap.String("source", event.Source),
			)
			return
		}

		switch res.name {
		case "liquidity":
			if res.err != nil {
				failReasons = append(failReasons, fmt.Sprintf("liquidity_error:%v", res.err))
				passed = false
			} else {
				liq := res.val.(float64)
				result.LiquidityUSD = liq
				if liq < e.cfg.MinLiquidityUSD {
					failReasons = append(failReasons, fmt.Sprintf("low_liquidity:$%.0f", liq))
					passed = false
				}
			}

		case "age":
			if res.err != nil {
				e.logger.Warn("age check failed", zap.Error(res.err))
				// non-fatal: age check failing doesn't block
			} else {
				age := res.val.(int64)
				result.AgeSeconds = age
				if age > e.cfg.MaxAgeSec {
					failReasons = append(failReasons, fmt.Sprintf("too_old:%ds", age))
					passed = false
				}
			}

		case "tokeninfo":
			if res.err == nil {
				info := res.val.(tokenInfoVal)
				tInfo = info
			}

		case "honeypot":
			if res.err != nil {
				e.logger.Warn("honeypot check error", zap.Error(res.err))
			} else {
				isHP := res.val.(bool)
				result.IsHoneypot = isHP
				if isHP {
					failReasons = append(failReasons, "honeypot_detected")
					passed = false
				}
			}
		}
	}

	// BSCscan holder check (best-effort, outside 2.5s window — fire-and-forget result)
	holderCount, top10Pct, rugScore := e.getBSCscanData(ctx, memeToken)
	result.HolderCount = holderCount
	result.Top10Pct = top10Pct
	result.RugScore = rugScore

	if holderCount > 0 {
		if holderCount < e.cfg.MinHolders {
			failReasons = append(failReasons, fmt.Sprintf("low_holders:%d", holderCount))
			passed = false
		}
		if top10Pct > e.cfg.MaxTop10Pct {
			failReasons = append(failReasons, fmt.Sprintf("top10_conc:%.1f%%", top10Pct))
			passed = false
		}
	}
	if rugScore > e.cfg.MaxRugScore {
		failReasons = append(failReasons, fmt.Sprintf("high_rug_score:%d", rugScore))
		passed = false
	}

	result.Passed = passed
	result.FailReasons = failReasons

	// Persist token metadata
	_ = e.db.UpsertToken(filterCtx, &db.Token{
		Address:     memeToken,
		Symbol:      tInfo.symbol,
		Name:        tInfo.name,
		Decimals:    tInfo.decimals,
		PairAddress: event.PairAddress,
		BlockNumber: int64(event.BlockNumber),
	})

	if dbErr := e.db.InsertFilterResult(filterCtx, result); dbErr != nil {
		e.logger.Error("insert filter result", zap.Error(dbErr))
	}

	if !passed {
		e.logger.Info("token rejected",
			zap.String("token", memeToken),
			zap.String("source", event.Source),
			zap.Strings("reasons", failReasons),
		)
		return
	}

	_ = e.db.IncrBotCounters(ctx, 0, 1, 0)

	e.logger.Info("token APPROVED",
		zap.String("token", memeToken),
		zap.String("symbol", tInfo.symbol),
		zap.String("pair", event.PairAddress),
		zap.String("source", event.Source),
		zap.Float64("liquidity_usd", result.LiquidityUSD),
	)

	approved := &ApprovedToken{
		TokenAddress: memeToken,
		PairAddress:  event.PairAddress,
		TokenSymbol:  tInfo.symbol,
		TokenName:    tInfo.name,
		Decimals:     tInfo.decimals,
		LiquidityUSD: result.LiquidityUSD,
		Source:       event.Source,
		BlockNumber:  event.BlockNumber,
		Timestamp:    time.Now().Unix(),
	}

	if err := e.redis.PublishToStream(ctx, redisclient.StreamApproved, approved); err != nil {
		e.logger.Error("publish approved token", zap.Error(err))
	}
	_ = e.redis.Publish(ctx, redisclient.PubSubEvents, map[string]interface{}{
		"type":  "token_approved",
		"token": approved,
	})
}

// ─── Individual checks ────────────────────────────────────────────────────────

func (e *Engine) getLiquidityUSD(ctx context.Context, pairAddress, memeToken string) (float64, error) {
	pairAddr := common.HexToAddress(pairAddress)

	e.acquireRPC()
	data, err := e.pairABI.Pack("getReserves")
	if err != nil {
		return 0, err
	}
	result, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &pairAddr, Data: data}, nil)
	if err != nil {
		return 0, fmt.Errorf("getReserves: %w", err)
	}

	type reserves struct {
		Reserve0           *big.Int
		Reserve1           *big.Int
		BlockTimestampLast uint32
	}
	var res reserves
	if err := e.pairABI.UnpackIntoInterface(&res, "getReserves", result); err != nil {
		return 0, fmt.Errorf("unpack reserves: %w", err)
	}
	if res.Reserve0.Sign() == 0 && res.Reserve1.Sign() == 0 {
		return 0, fmt.Errorf("zero reserves")
	}

	e.acquireRPC()
	data2, err := e.pairABI.Pack("token0")
	if err != nil {
		return 0, err
	}
	t0Result, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &pairAddr, Data: data2}, nil)
	if err != nil {
		return 0, fmt.Errorf("token0 call: %w", err)
	}
	var token0Addr common.Address
	if err := e.pairABI.UnpackIntoInterface(&token0Addr, "token0", t0Result); err != nil {
		return 0, fmt.Errorf("unpack token0: %w", err)
	}

	wbnb := common.HexToAddress(WBNBAddr)
	var bnbReserve *big.Int
	if token0Addr == wbnb {
		bnbReserve = res.Reserve0
	} else {
		bnbReserve = res.Reserve1
	}

	bnbPrice := e.getBNBPrice(ctx)
	bnbAmount, _ := new(big.Float).Quo(
		new(big.Float).SetInt(bnbReserve),
		new(big.Float).SetFloat64(1e18),
	).Float64()
	return bnbAmount * bnbPrice * 2, nil
}

func (e *Engine) getBNBPrice(ctx context.Context) float64 {
	bnbPriceMu.RLock()
	if time.Since(cachedBNBPriceAt) < 30*time.Second && cachedBNBPrice > 0 {
		p := cachedBNBPrice
		bnbPriceMu.RUnlock()
		return p
	}
	bnbPriceMu.RUnlock()

	// Try CoinGecko
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, "GET",
		"https://api.coingecko.com/api/v3/simple/price?ids=binancecoin&vs_currencies=usd", nil)
	if err != nil {
		return e.fallbackBNBPrice(ctx)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return e.fallbackBNBPrice(ctx)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var cg map[string]map[string]float64
	if err := json.Unmarshal(body, &cg); err != nil {
		return e.fallbackBNBPrice(ctx)
	}
	if price, ok := cg["binancecoin"]["usd"]; ok && price > 0 {
		bnbPriceMu.Lock()
		cachedBNBPrice = price
		cachedBNBPriceAt = time.Now()
		bnbPriceMu.Unlock()
		return price
	}
	return e.fallbackBNBPrice(ctx)
}

// fallbackBNBPrice reads WBNB/USDT pair reserves on-chain.
func (e *Engine) fallbackBNBPrice(ctx context.Context) float64 {
	// PancakeSwap WBNB-USDT pair
	const wbnbUsdtPair = "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE"
	pairAddr := common.HexToAddress(wbnbUsdtPair)
	data, err := e.pairABI.Pack("getReserves")
	if err != nil {
		return 300
	}
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	e.acquireRPC()
	result, err := e.rpc.CallContract(ctx2, ethereum.CallMsg{To: &pairAddr, Data: data}, nil)
	if err != nil {
		return 300
	}
	type reserves struct {
		Reserve0           *big.Int
		Reserve1           *big.Int
		BlockTimestampLast uint32
	}
	var res reserves
	if err := e.pairABI.UnpackIntoInterface(&res, "getReserves", result); err != nil {
		return 300
	}
	// WBNB is token0, USDT (6 decimals) is token1
	// price = usdt_reserve / wbnb_reserve * (1e18/1e6)
	if res.Reserve0.Sign() == 0 {
		return 300
	}
	usdtF, _ := new(big.Float).SetInt(res.Reserve1).Float64()
	wbnbF, _ := new(big.Float).SetInt(res.Reserve0).Float64()
	price := (usdtF / wbnbF) * 1e12 // adjust for decimal difference
	if price < 100 || price > 10000 {
		return 300
	}
	bnbPriceMu.Lock()
	cachedBNBPrice = price
	cachedBNBPriceAt = time.Now()
	bnbPriceMu.Unlock()
	return price
}

func (e *Engine) getTokenAge(ctx context.Context, pairBlock uint64) (int64, error) {
	e.acquireRPC()
	current, err := e.rpc.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}
	if current < pairBlock {
		return 0, nil
	}
	return int64(current-pairBlock) * 3, nil // BSC ~3s block time
}

type tokenInfoVal struct {
	symbol   string
	name     string
	decimals int
}

func (e *Engine) getTokenInfo(ctx context.Context, tokenAddress string) (tokenInfoVal, error) {
	addr := common.HexToAddress(tokenAddress)
	info := tokenInfoVal{decimals: 18}

	callAndUnpack := func(method string, out interface{}) error {
		data, err := e.erc20ABI.Pack(method)
		if err != nil {
			return err
		}
		e.acquireRPC()
		result, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &addr, Data: data}, nil)
		if err != nil {
			return err
		}
		return e.erc20ABI.UnpackIntoInterface(out, method, result)
	}

	var sym string
	if err := callAndUnpack("symbol", &sym); err == nil {
		info.symbol = sym
	}
	var name string
	if err := callAndUnpack("name", &name); err == nil {
		info.name = name
	}
	var dec uint8
	if err := callAndUnpack("decimals", &dec); err == nil {
		info.decimals = int(dec)
	}
	return info, nil
}

// checkHoneypot simulates buy then sell via getAmountsOut.
// Returns true if sell tax > 50% or sell simulation reverts.
func (e *Engine) checkHoneypot(ctx context.Context, tokenAddress, pairAddress string) (bool, error) {
	routerAddr := common.HexToAddress(RouterAddress)
	tokenAddr := common.HexToAddress(tokenAddress)
	wbnbAddr := common.HexToAddress(WBNBAddr)

	testAmountIn := big.NewInt(1e15) // 0.001 BNB
	buyPath := []common.Address{wbnbAddr, tokenAddr}

	buyData, err := e.routerABI.Pack("getAmountsOut", testAmountIn, buyPath)
	if err != nil {
		return false, err
	}
	e.acquireRPC()
	buyResult, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &routerAddr, Data: buyData}, nil)
	if err != nil {
		// Router call failed — could be new pair not yet in AMM
		return false, nil
	}
	var buyAmounts []*big.Int
	if err := e.routerABI.UnpackIntoInterface(&buyAmounts, "getAmountsOut", buyResult); err != nil {
		return true, nil
	}
	if len(buyAmounts) < 2 || buyAmounts[1].Sign() == 0 {
		return true, nil
	}
	tokensReceived := buyAmounts[1]

	// Simulate sell
	sellPath := []common.Address{tokenAddr, wbnbAddr}
	sellData, err := e.routerABI.Pack("getAmountsOut", tokensReceived, sellPath)
	if err != nil {
		return false, nil
	}
	e.acquireRPC()
	sellResult, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &routerAddr, Data: sellData}, nil)
	if err != nil {
		// Sell reverts → honeypot
		return true, nil
	}
	var sellAmounts []*big.Int
	if err := e.routerABI.UnpackIntoInterface(&sellAmounts, "getAmountsOut", sellResult); err != nil {
		return true, nil
	}
	if len(sellAmounts) < 2 || sellAmounts[1].Sign() == 0 {
		return true, nil
	}

	// Effective tax = 1 - (bnbBack / bnbIn)
	bnbBack := new(big.Float).SetInt(sellAmounts[1])
	bnbIn := new(big.Float).SetInt(testAmountIn)
	ratio, _ := new(big.Float).Quo(bnbBack, bnbIn).Float64()
	if ratio < 0.5 {
		return true, fmt.Errorf("sell_tax_%.0f%%", (1-ratio)*100)
	}
	return false, nil
}

func (e *Engine) getBSCscanData(ctx context.Context, tokenAddress string) (holderCount int, top10Pct float64, rugScore int) {
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	url := fmt.Sprintf(
		"https://api.bscscan.com/api?module=token&action=tokenholderlist&contractaddress=%s&page=1&offset=25&sort=desc",
		tokenAddress,
	)
	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 32768))
	type holder struct {
		TokenHolderAddress  string `json:"TokenHolderAddress"`
		TokenHolderQuantity string `json:"TokenHolderQuantity"`
	}
	var result struct {
		Status  string   `json:"status"`
		Result  []holder `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return
	}
	if result.Status != "1" || len(result.Result) == 0 {
		return
	}
	holderCount = len(result.Result)
	if holderCount < 25 {
		rugScore++
	}
	return
}
