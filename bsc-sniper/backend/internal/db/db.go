package db

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

//go:embed schema.sql
var schema string

type DB struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

func New(ctx context.Context, dsn string, logger *zap.Logger) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	d := &DB{Pool: pool, logger: logger}
	if err := d.migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	logger.Info("database connected and migrated")
	return d, nil
}

func (d *DB) migrate(ctx context.Context) error {
	_, err := d.Pool.Exec(ctx, schema)
	return err
}

func (d *DB) Close() {
	d.Pool.Close()
}

// ---- Token ----

type Token struct {
	ID          int64
	Address     string
	Symbol      string
	Name        string
	Decimals    int
	PairAddress string
	BlockNumber int64
	CreatedAt   time.Time
}

func (d *DB) UpsertToken(ctx context.Context, t *Token) error {
	_, err := d.Pool.Exec(ctx, `
		INSERT INTO tokens (address, symbol, name, decimals, pair_address, block_number)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (address) DO UPDATE SET
			symbol=EXCLUDED.symbol,
			name=EXCLUDED.name,
			decimals=EXCLUDED.decimals,
			pair_address=EXCLUDED.pair_address,
			updated_at=NOW()
	`, t.Address, t.Symbol, t.Name, t.Decimals, t.PairAddress, t.BlockNumber)
	return err
}

func (d *DB) ListTokens(ctx context.Context, limit int) ([]*Token, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT id, address, symbol, name, decimals, pair_address, block_number, created_at
		FROM tokens ORDER BY created_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*Token
	for rows.Next() {
		t := &Token{}
		if err := rows.Scan(&t.ID, &t.Address, &t.Symbol, &t.Name, &t.Decimals,
			&t.PairAddress, &t.BlockNumber, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// ---- Filter Result ----

type FilterResult struct {
	ID           int64
	TokenAddress string
	PairAddress  string
	LiquidityUSD float64
	AgeSeconds   int64
	HolderCount  int
	Top10Pct     float64
	RugScore     int
	IsHoneypot   bool
	Passed       bool
	FailReasons  []string
	CheckedAt    time.Time
}

func (d *DB) InsertFilterResult(ctx context.Context, f *FilterResult) error {
	_, err := d.Pool.Exec(ctx, `
		INSERT INTO filter_results
		  (token_address, pair_address, liquidity_usd, age_seconds, holder_count,
		   top10_pct, rug_score, is_honeypot, passed, fail_reasons)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, f.TokenAddress, f.PairAddress, f.LiquidityUSD, f.AgeSeconds, f.HolderCount,
		f.Top10Pct, f.RugScore, f.IsHoneypot, f.Passed, f.FailReasons)
	return err
}

// ---- Trade ----

type Trade struct {
	ID             int64
	TokenAddress   string
	PairAddress    string
	Side           string // buy | sell
	AmountBNB      float64
	AmountTokens   string
	PriceBNB       string
	TxHash         string
	GasUsed        int64
	GasPriceGwei   float64
	Status         string
	ErrorMsg       string
	ExecutedAt     time.Time
}

func (d *DB) InsertTrade(ctx context.Context, t *Trade) (int64, error) {
	var id int64
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO trades
		  (token_address, pair_address, side, amount_bnb, amount_tokens, price_bnb,
		   tx_hash, gas_used, gas_price_gwei, status, error_msg)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id
	`, t.TokenAddress, t.PairAddress, t.Side, t.AmountBNB, t.AmountTokens, t.PriceBNB,
		t.TxHash, t.GasUsed, t.GasPriceGwei, t.Status, t.ErrorMsg).Scan(&id)
	return id, err
}

func (d *DB) UpdateTradeStatus(ctx context.Context, id int64, status, txHash, errMsg string, gasUsed int64) error {
	_, err := d.Pool.Exec(ctx, `
		UPDATE trades SET status=$2, tx_hash=$3, error_msg=$4, gas_used=$5 WHERE id=$1
	`, id, status, txHash, errMsg, gasUsed)
	return err
}

func (d *DB) ListTrades(ctx context.Context, limit int) ([]*Trade, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT id, token_address, pair_address, side, amount_bnb, amount_tokens,
		       price_bnb, tx_hash, gas_used, gas_price_gwei, status, error_msg, executed_at
		FROM trades ORDER BY executed_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []*Trade
	for rows.Next() {
		t := &Trade{}
		if err := rows.Scan(&t.ID, &t.TokenAddress, &t.PairAddress, &t.Side,
			&t.AmountBNB, &t.AmountTokens, &t.PriceBNB, &t.TxHash, &t.GasUsed,
			&t.GasPriceGwei, &t.Status, &t.ErrorMsg, &t.ExecutedAt); err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}
	return trades, rows.Err()
}

// ---- Position ----

type Position struct {
	ID               int64
	TokenAddress     string
	PairAddress      string
	TokenSymbol      string
	QuoteToken       string // canonical-case address of the quote token
	EntryPriceBNB    string
	CurrentPriceBNB  string
	ATHPriceBNB      string
	AmountTokens     string
	CostBNB          float64
	RealizedPnlBNB   float64
	TP1Triggered     bool
	Status           string
	OpenedAt         time.Time
	ClosedAt         *time.Time
}

func (d *DB) UpsertPosition(ctx context.Context, p *Position) error {
	qt := p.QuoteToken
	if qt == "" {
		qt = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c" // default WBNB
	}
	_, err := d.Pool.Exec(ctx, `
		INSERT INTO positions
		  (token_address, pair_address, token_symbol, quote_token, entry_price_bnb, current_price_bnb,
		   ath_price_bnb, amount_tokens, cost_bnb, realized_pnl_bnb, tp1_triggered, status, closed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (token_address) DO UPDATE SET
		  current_price_bnb=EXCLUDED.current_price_bnb,
		  ath_price_bnb=EXCLUDED.ath_price_bnb,
		  amount_tokens=EXCLUDED.amount_tokens,
		  realized_pnl_bnb=EXCLUDED.realized_pnl_bnb,
		  tp1_triggered=EXCLUDED.tp1_triggered,
		  status=EXCLUDED.status,
		  closed_at=EXCLUDED.closed_at,
		  cost_bnb=EXCLUDED.cost_bnb
	`, p.TokenAddress, p.PairAddress, p.TokenSymbol, qt, p.EntryPriceBNB, p.CurrentPriceBNB,
		p.ATHPriceBNB, p.AmountTokens, p.CostBNB, p.RealizedPnlBNB, p.TP1Triggered, p.Status, p.ClosedAt)
	return err
}

func (d *DB) CountRecentPositionsBySymbol(ctx context.Context, symbol string, window time.Duration) (int64, error) {
	var count int64
	err := d.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM positions
		WHERE token_symbol = $1 AND opened_at > NOW() - $2::interval
	`, symbol, window).Scan(&count)
	return count, err
}

func (d *DB) ListPositions(ctx context.Context, status string) ([]*Position, error) {
	query := `
		SELECT id, token_address, pair_address, token_symbol,
		       COALESCE(quote_token, '0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095b'),
		       entry_price_bnb, current_price_bnb, ath_price_bnb, amount_tokens, cost_bnb,
		       realized_pnl_bnb, tp1_triggered, status, opened_at, closed_at
		FROM positions`
	args := []interface{}{}
	if status != "" {
		query += " WHERE status=$1"
		args = append(args, status)
	}
	query += " ORDER BY opened_at DESC"

	rows, err := d.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*Position
	for rows.Next() {
		p := &Position{}
		if err := rows.Scan(&p.ID, &p.TokenAddress, &p.PairAddress, &p.TokenSymbol,
			&p.QuoteToken,
			&p.EntryPriceBNB, &p.CurrentPriceBNB, &p.ATHPriceBNB, &p.AmountTokens,
			&p.CostBNB, &p.RealizedPnlBNB, &p.TP1Triggered, &p.Status,
			&p.OpenedAt, &p.ClosedAt); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, rows.Err()
}

// ---- Bot State ----

type BotState struct {
	Running      bool
	StartedAt    *time.Time
	StoppedAt    *time.Time
	PairsSeen    int64
	PairsPassed  int64
	TradesTotal  int64
	UpdatedAt    time.Time
}

func (d *DB) GetBotState(ctx context.Context) (*BotState, error) {
	s := &BotState{}
	err := d.Pool.QueryRow(ctx, `
		SELECT running, started_at, stopped_at, pairs_seen, pairs_passed, trades_total, updated_at
		FROM bot_state WHERE id=1
	`).Scan(&s.Running, &s.StartedAt, &s.StoppedAt, &s.PairsSeen, &s.PairsPassed, &s.TradesTotal, &s.UpdatedAt)
	return s, err
}

func (d *DB) SetBotRunning(ctx context.Context, running bool) error {
	if running {
		_, err := d.Pool.Exec(ctx, `
			UPDATE bot_state SET running=TRUE, started_at=NOW(), updated_at=NOW() WHERE id=1
		`)
		return err
	}
	_, err := d.Pool.Exec(ctx, `
		UPDATE bot_state SET running=FALSE, stopped_at=NOW(), updated_at=NOW() WHERE id=1
	`)
	return err
}

func (d *DB) IncrBotCounters(ctx context.Context, pairsSeen, pairsPassed, tradesTotal int64) error {
	_, err := d.Pool.Exec(ctx, `
		UPDATE bot_state SET
		  pairs_seen=pairs_seen+$1,
		  pairs_passed=pairs_passed+$2,
		  trades_total=trades_total+$3,
		  updated_at=NOW()
		WHERE id=1
	`, pairsSeen, pairsPassed, tradesTotal)
	return err
}
