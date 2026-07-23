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
  {"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
]`

const routerABIJSON = `[
  {"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"}
]`

const (
	RouterAddress = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
	WBNBAddr      = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
	wbnbAddrLower = "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c"

	// Elite safeguards
	eliteLiquidityFloorUSD = 12000.0 // $12k minimum pool liquidity for this test
	eliteMinRoundTripRatio = 0.95    // bnb back / bnb in; reject if efficiency < 95%

	// Stablecoin addresses (BSC, all 18 decimals)
	USDTAddr = "0x55d398326f99059fF775485246999027B3197955"
	BUSDAddr = "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"
	USDCAddr = "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d"

	// Variable-price quote tokens
	ETHAddr  = "0x2170Ed0880ac9A755fd29B2688956BD959F933F8"
	CAKEAddr = "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82"

	// On-chain reference pairs for fallback pricing
	wbnbUsdtPair = "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE"
	cakeWbnbPair = "0x0eD7e52944161450477ee417DE9Cd3a859b14fD0"
	ethWbnbPair  = "0x74E4716E431f45807DCF19f284c7aA99F18a4fbc"
)

// ─── Price caches (shared across all workers) ─────────────────────────────────

var (
	cachedBNBPrice   float64
	cachedBNBPriceAt time.Time
	bnbPriceMu       sync.RWMutex

	// Generic price cache for CAKE, ETH, etc.
	tokenPriceCache   = map[string]float64{}
	tokenPriceCacheAt = map[string]time.Time{}
	tokenPriceMu      sync.RWMutex
)

// ApprovedToken is the event published to stream:approved_tokens.
type ApprovedToken struct {
	TokenAddress string  `json:"token_address"`
	PairAddress  string  `json:"pair_address"`
	QuoteToken   string  `json:"quote_token"`  // canonical-case address
	QuoteSymbol  string  `json:"quote_symbol"` // "WBNB" | "USDT" | ...
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

	rpcTokens chan struct{}
	Active    atomic.Int64
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

	for i := 0; i < 100; i++ {
		e.rpcTokens <- struct{}{}
	}
	go e.refillRPCTokens()

	return e, nil
}

func (e *Engine) refillRPCTokens() {
	ticker := time.NewTicker(time.Second / 100)
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
	// Liquidity check retries up to 3×2s = 6s; other checks run in parallel.
	// Allow 15s total so the retry window fits without racing to a deadline.
	filterCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Resolve meme and quote tokens
	memeToken := event.MemeToken
	quoteToken := event.QuoteToken
	quoteSymbol := event.QuoteSymbol

	// Fallback: derive from token0/token1 if listener didn't populate
	if memeToken == "" || quoteToken == "" {
		t0 := strings.ToLower(event.Token0)
		t1 := strings.ToLower(event.Token1)
		if isQuoteToken(t0) && !isQuoteToken(t1) {
			memeToken = event.Token1
			quoteToken = event.Token0
		} else if isQuoteToken(t1) && !isQuoteToken(t0) {
			memeToken = event.Token0
			quoteToken = event.Token1
		} else {
			e.logger.Debug("pair has no recognised quote token, skipping",
				zap.String("t0", event.Token0),
				zap.String("t1", event.Token1),
			)
			return
		}
	}
	if quoteSymbol == "" {
		quoteSymbol = quoteSymbolFor(quoteToken)
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

	go func() {
		// Liquidity has its own internal retry loop; give it the full filterCtx.
		liq, err := e.getLiquidityUSD(filterCtx, event.PairAddress, memeToken, quoteToken)
		liquidityCh <- checkResult{"liquidity", err, liq}
	}()
	go func() {
		checkCtx, c := context.WithTimeout(filterCtx, 3*time.Second)
		defer c()
		age, err := e.getTokenAge(checkCtx, event.BlockNumber)
		ageCh <- checkResult{"age", err, age}
	}()
	go func() {
		checkCtx, c := context.WithTimeout(filterCtx, 3*time.Second)
		defer c()
		info, err := e.getTokenInfo(checkCtx, memeToken)
		tokenInfoCh <- checkResult{"tokeninfo", err, info}
	}()
	go func() {
		checkCtx, c := context.WithTimeout(filterCtx, 3*time.Second)
		defer c()
		hp, taxPct, efficiency, err := e.checkHoneypot(checkCtx, memeToken, event.PairAddress, quoteToken)
		honeypotCh <- checkResult{"honeypot", err, map[string]interface{}{"honeypot": hp, "tax_pct": taxPct, "efficiency": efficiency}}
	}()

	failReasons := []string{}
	passed := true
	var tInfo tokenInfoVal

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
				if liq < e.cfg.MinLiquidityUSD || liq < eliteLiquidityFloorUSD {
					failReasons = append(failReasons, fmt.Sprintf("low_liquidity:$%.0f", liq))
					passed = false
				}
			}

		case "age":
			if res.err != nil {
				e.logger.Warn("age check failed", zap.Error(res.err))
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
				tInfo = res.val.(tokenInfoVal)
			}

		case "honeypot":
			if res.err != nil {
				failReasons = append(failReasons, fmt.Sprintf("prebuy_sim_error:%v", res.err))
				passed = false
			} else {
				meta := res.val.(map[string]interface{})
				isHP := meta["honeypot"].(bool)
				taxPct := meta["tax_pct"].(float64)
				efficiency := meta["efficiency"].(float64)
				result.IsHoneypot = isHP
				// Log efficiency for EVERY token — the #1 metric for this test.
				e.logger.Info("prebuy efficiency",
					zap.String("token", memeToken),
					zap.String("symbol", tInfo.symbol),
					zap.String("quote", quoteSymbol),
					zap.Float64("efficiency", efficiency),
					zap.Float64("tax_pct", taxPct),
				)
				if isHP {
					failReasons = append(failReasons, "honeypot_detected")
					passed = false
				}
				if efficiency < eliteMinRoundTripRatio {
					failReasons = append(failReasons, fmt.Sprintf("low_efficiency:%.4f", efficiency))
					passed = false
				}
			}
		}
	}

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

	// Duplicate symbol guard: skip the same symbol if bought in the last 5 minutes.
	if e.isDuplicateSymbol(ctx, tInfo.symbol) {
		failReasons = append(failReasons, fmt.Sprintf("duplicate_symbol:%s", tInfo.symbol))
		passed = false
	}

	// USDT pair handling: warn, but do not auto-reject. The 2-hop simulation in
	// checkHoneypot already validates the WBNB/USDT conversion path.
	if strings.EqualFold(quoteToken, USDTAddr) {
		e.logger.Warn("USDT pair approved (warning only)",
			zap.String("token", memeToken),
			zap.String("symbol", tInfo.symbol),
		)
	}

	result.Passed = passed
	result.FailReasons = failReasons

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
			zap.String("quote", quoteSymbol),
			zap.String("source", event.Source),
			zap.Strings("reasons", failReasons),
		)
		return
	}

	_ = e.db.IncrBotCounters(ctx, 0, 1, 0)

	e.logger.Info("token APPROVED",
		zap.String("token", memeToken),
		zap.String("symbol", tInfo.symbol),
		zap.String("quote", quoteSymbol),
		zap.String("pair", event.PairAddress),
		zap.String("source", event.Source),
		zap.Float64("liquidity_usd", result.LiquidityUSD),
	)

	approved := &ApprovedToken{
		TokenAddress: memeToken,
		PairAddress:  event.PairAddress,
		QuoteToken:   quoteToken,
		QuoteSymbol:  quoteSymbol,
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

// ─── Liquidity calculation ────────────────────────────────────────────────────

// getLiquidityUSD returns total liquidity in USD for any supported quote token.
// Uses balanceOf(pool) on the quote ERC-20 — works for V2, V3, and StableSwap pools.
// Retries up to 3 times (2 s apart) when the pool has zero balance, because
// PairCreated fires before the creator adds initial liquidity.
func (e *Engine) getLiquidityUSD(ctx context.Context, pairAddress, memeToken, quoteToken string) (float64, error) {
	poolAddr := common.HexToAddress(pairAddress)
	quoteAddr := common.HexToAddress(quoteToken)

	data, err := e.erc20ABI.Pack("balanceOf", poolAddr)
	if err != nil {
		return 0, fmt.Errorf("pack balanceOf: %w", err)
	}

	const maxRetries = 3
	const retryDelay = 2 * time.Second

	var quoteBal *big.Int
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(retryDelay):
			}
		}
		e.acquireRPC()
		result, rErr := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &quoteAddr, Data: data}, nil)
		if rErr != nil {
			return 0, fmt.Errorf("balanceOf quote token: %w", rErr)
		}
		var bal *big.Int
		if uErr := e.erc20ABI.UnpackIntoInterface(&bal, "balanceOf", result); uErr != nil {
			return 0, fmt.Errorf("unpack balanceOf: %w", uErr)
		}
		if bal != nil && bal.Sign() > 0 {
			quoteBal = bal
			break
		}
		// Still zero — liquidity not yet added; retry unless last attempt
	}

	if quoteBal == nil || quoteBal.Sign() == 0 {
		return 0, fmt.Errorf("zero quote balance after %d retries", maxRetries)
	}

	// Get USD price for this quote token
	pricePerQuote, err := e.getQuoteTokenPriceUSD(ctx, quoteToken)
	if err != nil {
		return 0, fmt.Errorf("price fetch for %s: %w", quoteSymbolFor(quoteToken), err)
	}

	decimals := quoteTokenDecimals(quoteToken) // 1e18 for all whitelisted BSC tokens
	quoteF, _ := new(big.Float).Quo(
		new(big.Float).SetInt(quoteBal),
		new(big.Float).SetFloat64(decimals),
	).Float64()

	// Total pool liquidity ≈ 2× the quote side (AMM invariant: each side ≈ equal value)
	return quoteF * pricePerQuote * 2, nil
}

// getQuoteTokenPriceUSD returns the USD price for a given quote token.
// Results are cached for 30 seconds.
// Stablecoins return 1.0 immediately.
// WBNB uses the existing BNB price logic.
// CAKE and ETH use CoinGecko with on-chain fallback.
func (e *Engine) getQuoteTokenPriceUSD(ctx context.Context, quoteToken string) (float64, error) {
	lower := strings.ToLower(quoteToken)

	// Stablecoins: always $1
	switch lower {
	case strings.ToLower(USDTAddr),
		strings.ToLower(BUSDAddr),
		strings.ToLower(USDCAddr):
		return 1.0, nil
	}

	// WBNB: existing logic
	if lower == wbnbAddrLower {
		p := e.getBNBPrice(ctx)
		if p <= 0 {
			return 0, fmt.Errorf("could not determine BNB price")
		}
		return p, nil
	}

	// CAKE / ETH — check cache first
	tokenPriceMu.RLock()
	if t, ok := tokenPriceCacheAt[lower]; ok && time.Since(t) < 30*time.Second {
		p := tokenPriceCache[lower]
		tokenPriceMu.RUnlock()
		return p, nil
	}
	tokenPriceMu.RUnlock()

	// Try CoinGecko
	cgID := ""
	switch lower {
	case strings.ToLower(CAKEAddr):
		cgID = "pancakeswap-token"
	case strings.ToLower(ETHAddr):
		cgID = "ethereum"
	}

	if cgID != "" {
		reqCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(reqCtx, "GET",
			fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", cgID), nil)
		if err == nil {
			req.Header.Set("Accept", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
				var cg map[string]map[string]float64
				if json.Unmarshal(body, &cg) == nil {
					if price, ok := cg[cgID]["usd"]; ok && price > 0 {
						tokenPriceMu.Lock()
						tokenPriceCache[lower] = price
						tokenPriceCacheAt[lower] = time.Now()
						tokenPriceMu.Unlock()
						return price, nil
					}
				}
			}
		}
	}

	// On-chain fallback
	fallback, err := e.onChainTokenPrice(ctx, quoteToken)
	if err != nil {
		return 0, fmt.Errorf("on-chain price fallback: %w", err)
	}
	tokenPriceMu.Lock()
	tokenPriceCache[lower] = fallback
	tokenPriceCacheAt[lower] = time.Now()
	tokenPriceMu.Unlock()
	return fallback, nil
}

// onChainTokenPrice reads the token/WBNB pair reserves to derive USD price.
func (e *Engine) onChainTokenPrice(ctx context.Context, tokenAddr string) (float64, error) {
	lower := strings.ToLower(tokenAddr)

	refPair := ""
	quoteIsToken0 := false

	switch lower {
	case strings.ToLower(CAKEAddr):
		refPair = cakeWbnbPair
		quoteIsToken0 = true // CAKE is token0 in CAKE/WBNB pair
	case strings.ToLower(ETHAddr):
		refPair = ethWbnbPair
		quoteIsToken0 = true // ETH is token0 in ETH/WBNB pair
	default:
		return 0, fmt.Errorf("no reference pair for %s", tokenAddr)
	}

	pairAddr := common.HexToAddress(refPair)
	data, err := e.pairABI.Pack("getReserves")
	if err != nil {
		return 0, err
	}
	ctx2, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	e.acquireRPC()
	result, err := e.rpc.CallContract(ctx2, ethereum.CallMsg{To: &pairAddr, Data: data}, nil)
	if err != nil {
		return 0, err
	}
	type reserves struct {
		Reserve0           *big.Int
		Reserve1           *big.Int
		BlockTimestampLast uint32
	}
	var res reserves
	if err := e.pairABI.UnpackIntoInterface(&res, "getReserves", result); err != nil || res.Reserve0.Sign() == 0 || res.Reserve1.Sign() == 0 {
		return 0, fmt.Errorf("bad reserves for ref pair %s", refPair)
	}

	// price in BNB = wbnbReserve / tokenReserve
	var tokenRes, wbnbRes *big.Int
	if quoteIsToken0 {
		tokenRes = res.Reserve0
		wbnbRes = res.Reserve1
	} else {
		tokenRes = res.Reserve1
		wbnbRes = res.Reserve0
	}

	tokenF, _ := new(big.Float).SetInt(tokenRes).Float64()
	wbnbF, _ := new(big.Float).SetInt(wbnbRes).Float64()
	priceInBNB := wbnbF / tokenF

	bnbPrice := e.getBNBPrice(ctx)
	return priceInBNB * bnbPrice, nil
}

func (e *Engine) getBNBPrice(ctx context.Context) float64 {
	bnbPriceMu.RLock()
	if time.Since(cachedBNBPriceAt) < 30*time.Second && cachedBNBPrice > 0 {
		p := cachedBNBPrice
		bnbPriceMu.RUnlock()
		return p
	}
	bnbPriceMu.RUnlock()

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

func (e *Engine) fallbackBNBPrice(ctx context.Context) float64 {
	pairAddr := common.HexToAddress(wbnbUsdtPair)
	data, err := e.pairABI.Pack("getReserves")
	if err != nil {
		return 600
	}
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	e.acquireRPC()
	result, err := e.rpc.CallContract(ctx2, ethereum.CallMsg{To: &pairAddr, Data: data}, nil)
	if err != nil {
		return 600
	}
	type reserves struct {
		Reserve0           *big.Int
		Reserve1           *big.Int
		BlockTimestampLast uint32
	}
	var res reserves
	if err := e.pairABI.UnpackIntoInterface(&res, "getReserves", result); err != nil {
		return 600
	}
	if res.Reserve0.Sign() == 0 {
		return 600
	}
	// WBNB/USDT pair: both 18 decimals on BSC
	usdtF, _ := new(big.Float).SetInt(res.Reserve1).Float64()
	wbnbF, _ := new(big.Float).SetInt(res.Reserve0).Float64()
	price := usdtF / wbnbF
	if price < 100 || price > 10000 {
		return 600
	}
	bnbPriceMu.Lock()
	cachedBNBPrice = price
	cachedBNBPriceAt = time.Now()
	bnbPriceMu.Unlock()
	return price
}

// ─── Age check ────────────────────────────────────────────────────────────────

func (e *Engine) getTokenAge(ctx context.Context, pairBlock uint64) (int64, error) {
	e.acquireRPC()
	current, err := e.rpc.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}
	if current < pairBlock {
		return 0, nil
	}
	return int64(current-pairBlock) * 3, nil
}

// ─── Token info ───────────────────────────────────────────────────────────────

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

// ─── Honeypot simulation ──────────────────────────────────────────────────────

// checkHoneypot performs a static pre-buy simulation (eth_call) of the
// exact buy route followed by an immediate sell of 100% of the output.
// It returns (isBad, taxPct, efficiency, error):
//   - isBad      : true if the sell path reverts or returns zero (honeypot) or efficiency < 0.95.
//   - taxPct     : implied round-trip tax percentage (100 - efficiency*100).
//   - efficiency : simulated sell BNB / simulated buy BNB (the #1 metric).
//   - error     : non-nil if the simulation itself could not be completed.
func (e *Engine) checkHoneypot(ctx context.Context, tokenAddress, pairAddress, quoteToken string) (bool, float64, float64, error) {
	routerAddr := common.HexToAddress(RouterAddress)
	tokenAddr := common.HexToAddress(tokenAddress)
	wbnbAddr := common.HexToAddress(WBNBAddr)
	quoteAddr := common.HexToAddress(quoteToken)

	testAmountIn := big.NewInt(1e15) // 0.001 BNB — matches the live buy size

	// Build paths matching the actual executor route
	isWBNB := strings.EqualFold(quoteToken, WBNBAddr)
	var buyPath, sellPath []common.Address
	if isWBNB {
		buyPath = []common.Address{wbnbAddr, tokenAddr}
		sellPath = []common.Address{tokenAddr, wbnbAddr}
	} else {
		buyPath = []common.Address{wbnbAddr, quoteAddr, tokenAddr}
		sellPath = []common.Address{tokenAddr, quoteAddr, wbnbAddr}
	}

	// 1. Simulate buy: get expected token output for 0.001 BNB
	buyData, err := e.routerABI.Pack("getAmountsOut", testAmountIn, buyPath)
	if err != nil {
		return false, 0, 0, err
	}
	e.acquireRPC()
	buyResult, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &routerAddr, Data: buyData}, nil)
	if err != nil {
		// No direct route yet — not necessarily a honeypot, just not tradable.
		return false, 0, 0, nil
	}
	var buyAmounts []*big.Int
	if err := e.routerABI.UnpackIntoInterface(&buyAmounts, "getAmountsOut", buyResult); err != nil {
		return true, 0, 0, nil
	}
	if len(buyAmounts) < 2 || buyAmounts[len(buyAmounts)-1].Sign() == 0 {
		return true, 0, 0, nil
	}
	tokensReceived := buyAmounts[len(buyAmounts)-1]

	// 2. Simulate sell: immediately sell 100% of tokensReceived back to BNB
	sellData, err := e.routerABI.Pack("getAmountsOut", tokensReceived, sellPath)
	if err != nil {
		return false, 0, 0, nil
	}
	e.acquireRPC()
	sellResult, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &routerAddr, Data: sellData}, nil)
	if err != nil {
		// Sell reverts → honeypot
		return true, 0, 0, nil
	}
	var sellAmounts []*big.Int
	if err := e.routerABI.UnpackIntoInterface(&sellAmounts, "getAmountsOut", sellResult); err != nil {
		return true, 0, 0, nil
	}
	if len(sellAmounts) < 2 || sellAmounts[len(sellAmounts)-1].Sign() == 0 {
		return true, 0, 0, nil
	}

	// 3. Effective round-trip efficiency: bnbBack / bnbIn
	bnbBack := new(big.Float).SetInt(sellAmounts[len(sellAmounts)-1])
	bnbIn := new(big.Float).SetInt(testAmountIn)
	efficiency, _ := new(big.Float).Quo(bnbBack, bnbIn).Float64()
	taxPct := (1 - efficiency) * 100

	// Honeypot if sell reverts/returns zero OR round-trip efficiency is below 85%.
	isBad := efficiency < eliteMinRoundTripRatio
	return isBad, taxPct, efficiency, nil
}

// ─── BSCscan data ─────────────────────────────────────────────────────────────

func (e *Engine) getBSCscanData(ctx context.Context, tokenAddress string) (holderCount int, top10Pct float64, rugScore int) {
	url := fmt.Sprintf(
		"https://api.bscscan.com/api?module=token&action=tokenholderlist&contractaddress=%s&page=1&offset=25&sort=desc",
		tokenAddress,
	)

	type holder struct {
		TokenHolderAddress  string `json:"TokenHolderAddress"`
		TokenHolderQuantity string `json:"TokenHolderQuantity"`
	}
	var result struct {
		Status string   `json:"status"`
		Result []holder `json:"result"`
	}

	const maxRetries = 3
	const retryDelay = 500 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryDelay):
			}
		}

		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
		if err != nil {
			cancel()
			continue
		}
		req.Header.Set("Accept", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			cancel()
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 32768))
		resp.Body.Close()
		cancel()

		if err := json.Unmarshal(body, &result); err != nil {
			continue
		}
		if result.Status == "1" && len(result.Result) > 0 {
			break
		}
	}

	if result.Status != "1" || len(result.Result) == 0 {
		e.logger.Warn("BSCScan holder lookup failed after retries",
			zap.String("token", tokenAddress),
		)
		return
	}

	holderCount = len(result.Result)
	if holderCount < 25 {
		rugScore++
	}

	// Compute top-10 holder concentration. The BSCScan endpoint returns up to 25
	// holders sorted descending by quantity, so the first 10 are the largest.
	var top10Sum, totalSupply big.Float
	for i, h := range result.Result {
		qty, ok := new(big.Int).SetString(h.TokenHolderQuantity, 10)
		if !ok {
			continue
		}
		if i < 10 {
			top10Sum.Add(&top10Sum, new(big.Float).SetInt(qty))
		}
		totalSupply.Add(&totalSupply, new(big.Float).SetInt(qty))
	}
	if totalSupply.Sign() > 0 {
		ratio, _ := new(big.Float).Quo(&top10Sum, &totalSupply).Float64()
		top10Pct = ratio * 100
	}

	e.logger.Info("BSCScan holder data",
		zap.String("token", tokenAddress),
		zap.Int("holders", holderCount),
		zap.Float64("top10_pct", top10Pct),
	)
	return
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func isQuoteToken(addrLower string) bool {
	switch addrLower {
	case wbnbAddrLower,
		strings.ToLower(USDTAddr),
		strings.ToLower(BUSDAddr),
		strings.ToLower(USDCAddr),
		strings.ToLower(ETHAddr),
		strings.ToLower(CAKEAddr):
		return true
	}
	return false
}

func (e *Engine) isDuplicateSymbol(ctx context.Context, symbol string) bool {
	if symbol == "" {
		return false
	}
	count, err := e.db.CountRecentPositionsBySymbol(ctx, symbol, 5*time.Minute)
	if err != nil {
		e.logger.Warn("duplicate symbol check failed", zap.Error(err))
		return false
	}
	return count > 0
}

func quoteSymbolFor(addr string) string {
	switch strings.ToLower(addr) {
	case wbnbAddrLower:
		return "WBNB"
	case strings.ToLower(USDTAddr):
		return "USDT"
	case strings.ToLower(BUSDAddr):
		return "BUSD"
	case strings.ToLower(USDCAddr):
		return "USDC"
	case strings.ToLower(ETHAddr):
		return "ETH"
	case strings.ToLower(CAKEAddr):
		return "CAKE"
	}
	return "UNKNOWN"
}

// quoteTokenDecimals returns the big.Float divisor (10^decimals) for a quote token.
// All whitelisted tokens on BSC use 18 decimals.
func quoteTokenDecimals(_ string) float64 {
	return 1e18
}
