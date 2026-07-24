-- BSC Sniper Database Schema

CREATE TABLE IF NOT EXISTS tokens (
    id              BIGSERIAL PRIMARY KEY,
    address         VARCHAR(42) NOT NULL UNIQUE,
    symbol          VARCHAR(64),
    name            VARCHAR(256),
    decimals        INT DEFAULT 18,
    pair_address    VARCHAR(42),
    block_number    BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tokens_address ON tokens(address);
CREATE INDEX IF NOT EXISTS idx_tokens_created_at ON tokens(created_at DESC);

CREATE TABLE IF NOT EXISTS filter_results (
    id              BIGSERIAL PRIMARY KEY,
    token_address   VARCHAR(42) NOT NULL,
    pair_address    VARCHAR(42),
    liquidity_usd   NUMERIC(20,6),
    age_seconds     BIGINT,
    holder_count    INT,
    top10_pct       NUMERIC(6,2),
    rug_score       INT DEFAULT 0,
    is_honeypot     BOOLEAN DEFAULT FALSE,
    passed          BOOLEAN NOT NULL DEFAULT FALSE,
    fail_reasons    TEXT[],
    checked_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_filter_token ON filter_results(token_address);
CREATE INDEX IF NOT EXISTS idx_filter_passed ON filter_results(passed, checked_at DESC);

CREATE TABLE IF NOT EXISTS trades (
    id              BIGSERIAL PRIMARY KEY,
    token_address   VARCHAR(42) NOT NULL,
    pair_address    VARCHAR(42),
    side            VARCHAR(4) NOT NULL CHECK (side IN ('buy','sell')),
    amount_bnb      NUMERIC(20,8),
    amount_tokens   NUMERIC(40,0),
    price_bnb       NUMERIC(30,18),
    tx_hash         VARCHAR(66),
    gas_used        BIGINT,
    gas_price_gwei  NUMERIC(20,9),
    status          VARCHAR(16) DEFAULT 'pending' CHECK (status IN ('pending','confirmed','failed')),
    error_msg       TEXT,
    executed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trades_token ON trades(token_address);
CREATE INDEX IF NOT EXISTS idx_trades_executed_at ON trades(executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_trades_tx_hash ON trades(tx_hash);

CREATE TABLE IF NOT EXISTS positions (
    id              BIGSERIAL PRIMARY KEY,
    token_address   VARCHAR(42) NOT NULL UNIQUE,
    pair_address    VARCHAR(42),
    token_symbol    VARCHAR(64),
    quote_token     VARCHAR(42) DEFAULT '0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c',
    entry_price_bnb NUMERIC(30,18) NOT NULL,
    current_price_bnb NUMERIC(30,18),
    ath_price_bnb   NUMERIC(30,18),
    amount_tokens   NUMERIC(40,0) NOT NULL,
    cost_bnb        NUMERIC(20,8) NOT NULL,
    realized_pnl_bnb NUMERIC(20,8) DEFAULT 0,
    tp1_triggered   BOOLEAN DEFAULT FALSE,
    tp2_done        BOOLEAN DEFAULT FALSE,
    status          VARCHAR(16) DEFAULT 'open' CHECK (status IN ('open','closed','partial','bought','unsellable')),
    opened_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_positions_status ON positions(status, opened_at DESC);
CREATE INDEX IF NOT EXISTS idx_positions_token ON positions(token_address);

-- Migration: allow 'bought' and 'unsellable' statuses for ultra-low-latency test.
ALTER TABLE positions DROP CONSTRAINT IF EXISTS positions_status_check;
ALTER TABLE positions ADD CONSTRAINT positions_status_check CHECK (status IN ('open','closed','partial','bought','unsellable'));

CREATE TABLE IF NOT EXISTS bot_state (
    id          INT PRIMARY KEY DEFAULT 1,
    running     BOOLEAN NOT NULL DEFAULT FALSE,
    started_at  TIMESTAMPTZ,
    stopped_at  TIMESTAMPTZ,
    pairs_seen  BIGINT DEFAULT 0,
    pairs_passed BIGINT DEFAULT 0,
    trades_total BIGINT DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id = 1)
);

INSERT INTO bot_state (id, running) VALUES (1, FALSE)
ON CONFLICT (id) DO NOTHING;
