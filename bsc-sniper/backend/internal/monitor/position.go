package monitor

import (
	"context"
	"math/big"
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

const WBNBAddr = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095b"

// Seller is a subset of the executor used by monitor to trigger sells.
type Seller interface {
	ExecuteSell(ctx context.Context, pos *db.Position, pct int) error
}

type positionState struct {
	pos         *db.Position
	entryPrice  float64
	athPrice    float64
	tp1Hit      bool
	openedAt    time.Time
}

type Monitor struct {
	cfg      *config.Config
	rpc      *ethclient.Client
	redis    *redisclient.Client
	db       *db.DB
	seller   Seller
	logger   *zap.Logger
	pairABI  abi.ABI

	mu         sync.RWMutex
	positions  map[string]*positionState
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
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("monitor shutting down")
			return
		case <-ticker.C:
			m.tick(ctx)
		}
	}
}

func (m *Monitor) tick(ctx context.Context) {
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

		price, err := m.getCurrentPrice(ctx, pos.PairAddress, pos.TokenAddress)
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

		// Update ATH
		if price > state.athPrice {
			state.athPrice = price
		}

		// Calculate PnL
		entryPrice := state.entryPrice
		if entryPrice <= 0 {
			m.mu.Unlock()
			continue
		}
		pricePctGain := (price - entryPrice) / entryPrice * 100
		timeOpen := time.Since(state.openedAt)

		var action string
		var sellPct int

		switch {
		// Take-Profit 1: +100%
		case !state.tp1Hit && pricePctGain >= m.cfg.TakeProfit1Pct:
			action = "tp1"
			sellPct = 50
			state.tp1Hit = true
			pos.TP1Triggered = true

		// Trailing stop after TP1
		case state.tp1Hit:
			drawdown := (state.athPrice - price) / state.athPrice * 100
			if drawdown >= m.cfg.TrailingStopPct {
				action = "trailing_stop"
				sellPct = 100
			}

		// Time-based exit: >2h and <20% gain
		case timeOpen > 2*time.Hour && pricePctGain < 20:
			action = "time_exit"
			sellPct = 100
		}

		// Update current price in position
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

		if action != "" {
			m.logger.Info("exit triggered",
				zap.String("action", action),
				zap.String("token", tokenAddr),
				zap.String("symbol", pos.TokenSymbol),
				zap.Float64("price_bnb", price),
				zap.Float64("pnl_pct", pricePctGain),
				zap.Int("sell_pct", sellPct),
			)

			sellErr := m.seller.ExecuteSell(ctx, pos, sellPct)
			if sellErr != nil {
				m.logger.Error("sell failed", zap.Error(sellErr), zap.String("token", tokenAddr))
				continue
			}

			m.mu.Lock()
			if sellPct == 100 {
				pos.Status = "closed"
				now := time.Now()
				pos.ClosedAt = &now
				delete(m.positions, tokenAddr)
			} else {
				// partial: keep tracking remaining
				state.pos.TP1Triggered = true
			}
			m.mu.Unlock()

			_ = m.db.UpsertPosition(ctx, pos)
		}
	}
}

func (m *Monitor) getCurrentPrice(ctx context.Context, pairAddress, tokenAddress string) (float64, error) {
	pairAddr := common.HexToAddress(pairAddress)
	priceCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// getReserves
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

	// token0
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

	var bnbReserve, tokenReserve *big.Int
	if token0Addr == wbnb {
		bnbReserve = res.Reserve0
		tokenReserve = res.Reserve1
	} else if token0Addr == token {
		tokenReserve = res.Reserve0
		bnbReserve = res.Reserve1
	} else {
		// Neither is WBNB — unexpected
		return 0, nil
	}

	if tokenReserve.Sign() == 0 {
		return 0, nil
	}

	// price = bnbReserve / tokenReserve (both in their native decimals, assume token has 18)
	price, _ := new(big.Float).Quo(
		new(big.Float).SetInt(bnbReserve),
		new(big.Float).SetInt(tokenReserve),
	).Float64()

	return price, nil
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
