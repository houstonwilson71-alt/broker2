package monitor

import (
	"context"
	"math/big"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/bsc-sniper/backend/internal/config"
	"github.com/bsc-sniper/backend/internal/db"
	redisclient "github.com/bsc-sniper/backend/internal/redis"
	"go.uber.org/zap"
)

const pairReservesABIJSON = `[
  {"inputs":[],"name":"getReserves","outputs":[{"internalType":"uint112","name":"_reserve0","type":"uint112"},{"internalType":"uint112","name":"_reserve1","type":"uint112"},{"internalType":"uint32","name":"_blockTimestampLast","type":"uint32"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}
]`

const (
	WBNBAddr    = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
	wbnbLower   = "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c"
	// WBNB/USDT pair for BNB price fallback
	wbnbUsdtRef = "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE"
)

// Seller is a subset of the executor used by monitor to trigger sells.
type Seller interface {
	ExecuteSell(ctx context.Context, pos *db.Position, pct int) error
}

type positionState struct {
	pos        *db.Position
	entryPrice float64
	athPrice   float64
	tp1Hit     bool
	openedAt   time.Time
}

type Monitor struct {
	cfg     *config.Config
	rpc     *ethclient.Client
	redis   *redisclient.Client
	db      *db.DB
	seller  Seller
	logger  *zap.Logger
	pairABI abi.ABI

	mu        sync.RWMutex
	positions map[string]*positionState
}

func New(cfg *config.Config, rpc *ethclient.Client, redis *redisclient.Client,
	database *db.DB, seller Seller, logger *zap.Logger) (*Monitor, error) {

	pairABI, err := abi.JSON(strings.NewReader(pairReservesABIJSON))
	if err != nil {
		return nil, err
	}

	m := &Monitor{
		cfg:       cfg,
		rpc:       rpc,
		redis:     redis,
		db:        database,
		seller:    seller,
		logger:    logger,
		pairABI:   pairABI,
		positions: make(map[string]*positionState),
	}
	return m, nil
}

func (m *Monitor) LoadFromDB(ctx context.Context) error {
	positions, err := m.db.ListPositions(ctx, "open")
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range positions {
		entry := parseBigFloat(p.EntryPriceBNB)
		ath := parseBigFloat(p.ATHPriceBNB)
		if ath == 0 {
			ath = entry
		}
		m.positions[p.TokenAddress] = &positionState{
			pos:        p,
			entryPrice: entry,
			athPrice:   ath,
			tp1Hit:     p.TP1Triggered,
			openedAt:   p.OpenedAt,
		}
	}
	m.logger.Info("loaded positions from DB", zap.Int("count", len(m.positions)))
	return nil
}

func (m *Monitor) AddPosition(pos *db.Position) {
	entry := parseBigFloat(pos.EntryPriceBNB)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.positions[pos.TokenAddress] = &positionState{
		pos:        pos,
		entryPrice: entry,
		athPrice:   entry,
		tp1Hit:     false,
		openedAt:   pos.OpenedAt,
	}
}

func (m *Monitor) Run(ctx context.Context) {
	priceTicker := time.NewTicker(3 * time.Second)
	defer priceTicker.Stop()
	reloadTicker := time.NewTicker(10 * time.Second)
	defer reloadTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("monitor shutting down — persisting positions")
			m.persistAllPositions()
			return
		case <-priceTicker.C:
			m.tick(ctx)
		case <-reloadTicker.C:
			m.reloadNewPositions(ctx)
		}
	}
}

// reloadNewPositions periodically queries the DB for open positions that are
// not yet tracked in memory. This catches positions created by the executor
// after the monitor started.
func (m *Monitor) reloadNewPositions(ctx context.Context) {
	positions, err := m.db.ListPositions(ctx, "open")
	if err != nil {
		m.logger.Warn("reload positions from DB", zap.Error(err))
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	added := 0
	for _, p := range positions {
		if _, ok := m.positions[p.TokenAddress]; ok {
			continue
		}
		entry := parseBigFloat(p.EntryPriceBNB)
		ath := parseBigFloat(p.ATHPriceBNB)
		if ath == 0 {
			ath = entry
		}
		m.positions[p.TokenAddress] = &positionState{
			pos:        p,
			entryPrice: entry,
			athPrice:   ath,
			tp1Hit:     p.TP1Triggered,
			openedAt:   p.OpenedAt,
		}
		added++
	}
	if added > 0 {
		m.logger.Info("reloaded new positions from DB", zap.Int("added", added))
	}
}

// persistAllPositions flushes all in-memory positions to DB on graceful shutdown.
func (m *Monitor) persistAllPositions() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, state := range m.positions {
		pos := state.pos
		pos.ATHPriceBNB = bigFloatStr(state.athPrice)
		pos.TP1Triggered = state.tp1Hit
		if err := m.db.UpsertPosition(ctx, pos); err != nil {
			m.logger.Error("persist position on shutdown", zap.Error(err), zap.String("token", pos.TokenAddress))
		}
	}
	m.logger.Info("positions persisted", zap.Int("count", len(m.positions)))
}

func (m *Monitor) tick(ctx context.Context) {
	// Panic recovery — never let a single bad position crash the monitor
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("monitor panic recovered",
				zap.Any("panic", r),
				zap.ByteString("stack", debug.Stack()),
			)
		}
	}()

	m.mu.RLock()
	tokens := make([]string, 0, len(m.positions))
	for t := range m.positions {
		tokens = append(tokens, t)
	}
	m.mu.RUnlock()

	for _, tokenAddr := range tokens {
		m.mu.RLock()
		state, ok := m.positions[tokenAddr]
		if !ok {
			m.mu.RUnlock()
			continue
		}
		pos := state.pos
		m.mu.RUnlock()

		price, err := m.getCurrentPrice(ctx, pos.PairAddress, pos.TokenAddress, pos.QuoteToken)
		if err != nil {
			continue
		}
		if price <= 0 {
			continue
		}

		m.mu.Lock()
		state, ok = m.positions[tokenAddr]
		if !ok {
			m.mu.Unlock()
			continue
		}

		entryPrice := state.entryPrice
		if entryPrice <= 0 {
			m.mu.Unlock()
			continue
		}
		pricePctGain := (price - entryPrice) / entryPrice * 100

		var action string
		var sellPct int

		switch {
		// TP at +200%: price has tripled relative to entry. Sell 100% of the position.
		case price >= state.entryPrice*3.0:
			action = "tp_200"
			sellPct = 100

		// Break-even stop-loss: if price drops to or below entry price, dump 100%.
		case price <= state.entryPrice:
			action = "breakeven_sl"
			sellPct = 100
		}

		pos.CurrentPriceBNB = bigFloatStr(price)
		pos.ATHPriceBNB = bigFloatStr(state.athPrice)
		m.mu.Unlock()

		// Persist price update
		_ = m.db.UpsertPosition(ctx, pos)

		// Publish price update event
		_ = m.redis.Publish(ctx, redisclient.PubSubEvents, map[string]interface{}{
			"type":          "price_update",
			"token":         tokenAddr,
			"symbol":        pos.TokenSymbol,
			"price_bnb":     price,
			"pnl_pct":       pricePctGain,
			"ath_price_bnb": state.athPrice,
		})

		if action == "" {
			continue
		}

		var msg string
		switch action {
		case "tp_200":
			msg = "TP 200% hit, sold all"
		case "breakeven_sl":
			msg = "Break-even SL triggered, sold all"
		}
		m.logger.Info(msg,
			zap.String("action", action),
			zap.String("token", tokenAddr),
			zap.String("symbol", pos.TokenSymbol),
			zap.Float64("price_bnb", price),
			zap.Float64("pnl_pct", pricePctGain),
		)

		sellErr := m.seller.ExecuteSell(ctx, pos, sellPct)
		if sellErr != nil {
			m.logger.Error("sell failed", zap.Error(sellErr), zap.String("token", tokenAddr))
			continue
		}

		m.mu.Lock()
		pos.Status = "closed"
		now := time.Now()
		pos.ClosedAt = &now
		delete(m.positions, tokenAddr)
		m.mu.Unlock()
		_ = m.db.UpsertPosition(ctx, pos)
	}
}

// getCurrentPrice returns the token price denominated in BNB.
// For WBNB-paired pools this is direct. For stablecoin/other pools it
// converts via the BNB price so all positions use the same unit.
func (m *Monitor) getCurrentPrice(ctx context.Context, pairAddress, tokenAddress, quoteToken string) (float64, error) {
	pairAddr := common.HexToAddress(pairAddress)
	priceCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	data, err := m.pairABI.Pack("getReserves")
	if err != nil {
		return 0, err
	}
	result, err := m.rpc.CallContract(priceCtx, ethereum.CallMsg{To: &pairAddr, Data: data}, nil)
	if err != nil {
		return 0, err
	}

	type reserves struct {
		Reserve0           *big.Int
		Reserve1           *big.Int
		BlockTimestampLast uint32
	}
	var res reserves
	if err := m.pairABI.UnpackIntoInterface(&res, "getReserves", result); err != nil {
		return 0, err
	}

	data2, err := m.pairABI.Pack("token0")
	if err != nil {
		return 0, err
	}
	t0Result, err := m.rpc.CallContract(priceCtx, ethereum.CallMsg{To: &pairAddr, Data: data2}, nil)
	if err != nil {
		return 0, err
	}
	var token0Addr common.Address
	if err := m.pairABI.UnpackIntoInterface(&token0Addr, "token0", t0Result); err != nil {
		return 0, err
	}

	wbnb := common.HexToAddress(WBNBAddr)
	token := common.HexToAddress(tokenAddress)

	// Determine which reserve is which
	var quoteReserve, tokenReserve *big.Int
	quoteAddr := wbnb
	if quoteToken != "" && !strings.EqualFold(quoteToken, WBNBAddr) {
		quoteAddr = common.HexToAddress(quoteToken)
	}

	if token0Addr == quoteAddr {
		quoteReserve = res.Reserve0
		tokenReserve = res.Reserve1
	} else if token0Addr == token {
		tokenReserve = res.Reserve0
		quoteReserve = res.Reserve1
	} else {
		return 0, nil
	}

	if tokenReserve == nil || tokenReserve.Sign() == 0 || quoteReserve == nil || quoteReserve.Sign() == 0 {
		return 0, nil
	}

	// price in quote-token units per meme token
	priceInQuote, _ := new(big.Float).Quo(
		new(big.Float).SetInt(quoteReserve),
		new(big.Float).SetInt(tokenReserve),
	).Float64()

	// Convert to BNB if the quote token is not WBNB
	if strings.EqualFold(quoteToken, WBNBAddr) || quoteToken == "" {
		return priceInQuote, nil
	}

	// For stablecoins: price_in_BNB = price_in_USD / BNB_price_USD
	bnbPrice := m.getBNBPrice(ctx)
	if bnbPrice <= 0 {
		return 0, nil
	}
	return priceInQuote / bnbPrice, nil
}

// getBNBPrice fetches BNB/USD price from the WBNB/USDT reference pair.
func (m *Monitor) getBNBPrice(ctx context.Context) float64 {
	pairAddr := common.HexToAddress(wbnbUsdtRef)
	pCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	data, err := m.pairABI.Pack("getReserves")
	if err != nil {
		return 600
	}
	result, err := m.rpc.CallContract(pCtx, ethereum.CallMsg{To: &pairAddr, Data: data}, nil)
	if err != nil {
		return 600
	}
	type reserves struct {
		Reserve0           *big.Int
		Reserve1           *big.Int
		BlockTimestampLast uint32
	}
	var res reserves
	if err := m.pairABI.UnpackIntoInterface(&res, "getReserves", result); err != nil || res.Reserve0.Sign() == 0 {
		return 600
	}
	// WBNB is token0, USDT is token1 — both 18 decimals on BSC
	usdtF, _ := new(big.Float).SetInt(res.Reserve1).Float64()
	wbnbF, _ := new(big.Float).SetInt(res.Reserve0).Float64()
	price := usdtF / wbnbF
	if price < 100 || price > 10000 {
		return 600
	}
	return price
}

func (m *Monitor) GetPositions() []*positionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*positionState, 0, len(m.positions))
	for _, s := range m.positions {
		result = append(result, s)
	}
	return result
}

// SellAllPositions force-sells every open position at 100%. Used before the bot stops.
func (m *Monitor) SellAllPositions(ctx context.Context) {
	m.mu.RLock()
	tokens := make([]string, 0, len(m.positions))
	for t := range m.positions {
		tokens = append(tokens, t)
	}
	m.mu.RUnlock()

	if len(tokens) == 0 {
		return
	}

	m.logger.Warn("FORCE SELL ALL — selling every open position", zap.Int("count", len(tokens)))

	for _, tokenAddr := range tokens {
		m.mu.RLock()
		state, ok := m.positions[tokenAddr]
		m.mu.RUnlock()
		if !ok {
			continue
		}
		pos := state.pos

		if err := m.seller.ExecuteSell(ctx, pos, 100); err != nil {
			m.logger.Error("force sell failed", zap.Error(err), zap.String("token", tokenAddr))
			continue
		}

		m.mu.Lock()
		if st, ok := m.positions[tokenAddr]; ok {
			st.pos.Status = "closed"
			now := time.Now()
			st.pos.ClosedAt = &now
			delete(m.positions, tokenAddr)
		}
		m.mu.Unlock()
		_ = m.db.UpsertPosition(ctx, pos)
	}
}

func parseBigFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _, _ := new(big.Float).Parse(s, 10)
	v, _ := f.Float64()
	return v
}

func bigFloatStr(f float64) string {
	return new(big.Float).SetFloat64(f).Text('f', 18)
}
