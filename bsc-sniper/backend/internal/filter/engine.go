package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
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

// PancakeSwap V2 Router
const RouterAddress = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
const WBNBAddr = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095b"

type ApprovedToken struct {
	TokenAddress string  `json:"token_address"`
	PairAddress  string  `json:"pair_address"`
	TokenSymbol  string  `json:"token_symbol"`
	TokenName    string  `json:"token_name"`
	Decimals     int     `json:"decimals"`
	LiquidityUSD float64 `json:"liquidity_usd"`
	BlockNumber  uint64  `json:"block_number"`
	Timestamp    int64   `json:"timestamp"`
}

type Engine struct {
	cfg    *config.Config
	rpc    *ethclient.Client
	redis  *redisclient.Client
	db     *db.DB
	logger *zap.Logger

	pairABI   abi.ABI
	erc20ABI  abi.ABI
	routerABI abi.ABI
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
	return &Engine{
		cfg:       cfg,
		rpc:       rpc,
		redis:     redis,
		db:        database,
		logger:    logger,
		pairABI:   pairABI,
		erc20ABI:  erc20ABI,
		routerABI: routerABI,
	}, nil
}

func (e *Engine) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			e.logger.Info("filter engine shutting down")
			return
		default:
		}

		streams, err := e.redis.ReadStream(ctx, redisclient.StreamNewPairs, redisclient.GroupFilter, redisclient.ConsumerFilter, 10, 2*time.Second)
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

				go func(msgID string, ev listener.NewPairEvent) {
					e.processEvent(ctx, &ev)
					_ = e.redis.AckMessage(ctx, redisclient.StreamNewPairs, redisclient.GroupFilter, msgID)
				}(msg.ID, event)

				_ = e.db.IncrBotCounters(ctx, 1, 0, 0)
			}
		}
	}
}

func (e *Engine) processEvent(ctx context.Context, event *listener.NewPairEvent) {
	filterCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Determine meme token address
	wbnb := strings.ToLower(WBNBAddr)
	memeToken := event.Token0
	if strings.ToLower(event.Token0) == wbnb {
		memeToken = event.Token1
	}

	result := &db.FilterResult{
		TokenAddress: memeToken,
		PairAddress:  event.PairAddress,
	}

	failReasons := []string{}
	passed := true

	// Run checks in parallel
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
		liq, err := e.getLiquidityUSD(filterCtx, event.PairAddress, memeToken)
		liquidityCh <- checkResult{"liquidity", err, liq}
	}()
	go func() {
		age, err := e.getTokenAge(filterCtx, event.BlockNumber)
		ageCh <- checkResult{"age", err, age}
	}()
	go func() {
		info, err := e.getTokenInfo(filterCtx, memeToken)
		tokenInfoCh <- checkResult{"tokeninfo", err, info}
	}()
	go func() {
		hp, err := e.checkHoneypot(filterCtx, memeToken, event.PairAddress)
		honeypotCh <- checkResult{"honeypot", err, hp}
	}()

	type tokenInfo struct {
		symbol   string
		name     string
		decimals int
	}
	var tInfo tokenInfo

	// Collect results
	for i := 0; i < 4; i++ {
		var res checkResult
		select {
		case res = <-liquidityCh:
		case res = <-ageCh:
		case res = <-tokenInfoCh:
		case res = <-honeypotCh:
		case <-filterCtx.Done():
			e.logger.Warn("filter timeout", zap.String("pair", event.PairAddress))
			return
		}

		switch res.name {
		case "liquidity":
			if res.err != nil {
				e.logger.Warn("liquidity check failed", zap.Error(res.err))
				failReasons = append(failReasons, "liquidity_check_failed")
				passed = false
			} else {
				liq := res.val.(float64)
				result.LiquidityUSD = liq
				if liq < e.cfg.MinLiquidityUSD {
					failReasons = append(failReasons, fmt.Sprintf("low_liquidity:%.2f", liq))
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
				info := res.val.(tokenInfo)
				tInfo = info
			}

		case "honeypot":
			if res.err != nil {
				e.logger.Warn("honeypot check error", zap.Error(res.err))
			} else {
				isHP := res.val.(bool)
				result.IsHoneypot = isHP
				if isHP {
					failReasons = append(failReasons, "honeypot")
					passed = false
				}
			}
		}
	}

	// BSCscan holder check (best-effort)
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
			failReasons = append(failReasons, fmt.Sprintf("top10_concentration:%.1f%%", top10Pct))
			passed = false
		}
	}
	if rugScore > e.cfg.MaxRugScore {
		failReasons = append(failReasons, fmt.Sprintf("high_rug_score:%d", rugScore))
		passed = false
	}

	result.Passed = passed
	result.FailReasons = failReasons

	// Upsert token
	_ = e.db.UpsertToken(filterCtx, &db.Token{
		Address:     memeToken,
		Symbol:      tInfo.symbol,
		Name:        tInfo.name,
		Decimals:    tInfo.decimals,
		PairAddress: event.PairAddress,
		BlockNumber: int64(event.BlockNumber),
	})

	// Insert filter result
	if dbErr := e.db.InsertFilterResult(filterCtx, result); dbErr != nil {
		e.logger.Error("insert filter result", zap.Error(dbErr))
	}

	if !passed {
		e.logger.Info("token rejected",
			zap.String("token", memeToken),
			zap.Strings("reasons", failReasons),
		)
		return
	}

	_ = e.db.IncrBotCounters(ctx, 0, 1, 0)

	e.logger.Info("token approved",
		zap.String("token", memeToken),
		zap.String("symbol", tInfo.symbol),
		zap.String("pair", event.PairAddress),
		zap.Float64("liquidity_usd", result.LiquidityUSD),
	)

	approved := &ApprovedToken{
		TokenAddress: memeToken,
		PairAddress:  event.PairAddress,
		TokenSymbol:  tInfo.symbol,
		TokenName:    tInfo.name,
		Decimals:     tInfo.decimals,
		LiquidityUSD: result.LiquidityUSD,
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

func (e *Engine) getLiquidityUSD(ctx context.Context, pairAddress, memeToken string) (float64, error) {
	pairAddr := common.HexToAddress(pairAddress)

	// getReserves
	data, err := e.pairABI.Pack("getReserves")
	if err != nil {
		return 0, err
	}
	result, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &pairAddr, Data: data}, nil)
	if err != nil {
		return 0, err
	}

	type reserves struct {
		Reserve0             *big.Int
		Reserve1             *big.Int
		BlockTimestampLast   uint32
	}
	var res reserves
	if err := e.pairABI.UnpackIntoInterface(&res, "getReserves", result); err != nil {
		return 0, err
	}

	// Get token0 to know which reserve is BNB
	data, err = e.pairABI.Pack("token0")
	if err != nil {
		return 0, err
	}
	t0Result, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &pairAddr, Data: data}, nil)
	if err != nil {
		return 0, err
	}
	var token0Addr common.Address
	if err := e.pairABI.UnpackIntoInterface(&token0Addr, "token0", t0Result); err != nil {
		return 0, err
	}

	wbnb := common.HexToAddress(WBNBAddr)
	var bnbReserve *big.Int
	if token0Addr == wbnb {
		bnbReserve = res.Reserve0
	} else {
		bnbReserve = res.Reserve1
	}

	// Get BNB price in USD via CoinGecko (cached)
	bnbPrice := e.getBNBPrice(ctx)

	// bnbReserve is in wei (1e18)
	bnbAmount := new(big.Float).Quo(
		new(big.Float).SetInt(bnbReserve),
		new(big.Float).SetFloat64(1e18),
	)
	bnbF, _ := bnbAmount.Float64()
	return bnbF * bnbPrice * 2, nil // *2 for both sides of liquidity
}

var cachedBNBPrice float64
var cachedBNBPriceAt time.Time

func (e *Engine) getBNBPrice(ctx context.Context) float64 {
	if time.Since(cachedBNBPriceAt) < 60*time.Second && cachedBNBPrice > 0 {
		return cachedBNBPrice
	}

	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET",
		"https://api.coingecko.com/api/v3/simple/price?ids=binancecoin&vs_currencies=usd", nil)
	if err != nil {
		return 300 // fallback
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 300
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return 300
	}

	var result map[string]map[string]float64
	if err := json.Unmarshal(body, &result); err != nil {
		return 300
	}
	if price, ok := result["binancecoin"]["usd"]; ok && price > 0 {
		cachedBNBPrice = price
		cachedBNBPriceAt = time.Now()
		return price
	}
	return 300
}

func (e *Engine) getTokenAge(ctx context.Context, pairBlock uint64) (int64, error) {
	current, err := e.rpc.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}
	// BSC block time ~3 seconds
	ageSec := int64(current-pairBlock) * 3
	return ageSec, nil
}

type tokenInfo struct {
	symbol   string
	name     string
	decimals int
}

func (e *Engine) getTokenInfo(ctx context.Context, tokenAddress string) (tokenInfo, error) {
	addr := common.HexToAddress(tokenAddress)
	info := tokenInfo{decimals: 18}

	callAndUnpack := func(method string, out interface{}) error {
		data, err := e.erc20ABI.Pack(method)
		if err != nil {
			return err
		}
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

func (e *Engine) checkHoneypot(ctx context.Context, tokenAddress, pairAddress string) (bool, error) {
	routerAddr := common.HexToAddress(RouterAddress)
	tokenAddr := common.HexToAddress(tokenAddress)
	wbnbAddr := common.HexToAddress(WBNBAddr)

	// Try to simulate getAmountsOut for buy (BNB -> token)
	testAmountIn := new(big.Int).Mul(big.NewInt(1e15), big.NewInt(1)) // 0.001 BNB
	path := []common.Address{wbnbAddr, tokenAddr}

	data, err := e.routerABI.Pack("getAmountsOut", testAmountIn, path)
	if err != nil {
		return false, err
	}

	result, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &routerAddr, Data: data}, nil)
	if err != nil {
		// Can't simulate → treat as suspicious but not definitive
		return false, nil
	}

	var amounts []*big.Int
	if err := e.routerABI.UnpackIntoInterface(&amounts, "getAmountsOut", result); err != nil {
		return true, nil
	}
	if len(amounts) < 2 || amounts[1].Sign() == 0 {
		return true, nil
	}

	// Simulate reverse (token -> BNB) to detect sell tax
	tokensOut := amounts[1]
	sellPath := []common.Address{tokenAddr, wbnbAddr}
	data2, err := e.routerABI.Pack("getAmountsOut", tokensOut, sellPath)
	if err != nil {
		return false, nil
	}
	result2, err := e.rpc.CallContract(ctx, ethereum.CallMsg{To: &routerAddr, Data: data2}, nil)
	if err != nil {
		// sell simulation fails → honeypot
		return true, nil
	}

	var amounts2 []*big.Int
	if err := e.routerABI.UnpackIntoInterface(&amounts2, "getAmountsOut", result2); err != nil {
		return true, nil
	}
	if len(amounts2) < 2 || amounts2[1].Sign() == 0 {
		return true, nil
	}

	// If we get back less than 50% of BNB → sell tax >50% → honeypot
	bnbBack := new(big.Float).SetInt(amounts2[1])
	bnbIn := new(big.Float).SetInt(testAmountIn)
	ratio, _ := new(big.Float).Quo(bnbBack, bnbIn).Float64()
	if ratio < 0.5 {
		return true, nil
	}

	return false, nil
}

// getBSCscanData fetches holder count and top-10 concentration from BSCscan API (best-effort).
func (e *Engine) getBSCscanData(ctx context.Context, tokenAddress string) (holderCount int, top10Pct float64, rugScore int) {
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.bscscan.com/api?module=token&action=tokenholderlist&contractaddress=%s&page=1&offset=25&sort=desc", tokenAddress)
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

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32768))
	if err != nil {
		return
	}

	type holder struct {
		TokenHolderAddress  string `json:"TokenHolderAddress"`
		TokenHolderQuantity string `json:"TokenHolderQuantity"`
	}
	var result struct {
		Status  string   `json:"status"`
		Message string   `json:"message"`
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
