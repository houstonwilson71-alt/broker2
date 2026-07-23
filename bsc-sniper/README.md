# BSC Sniper — Production Meme Coin Sniper for BNB Smart Chain

A fully-automated, production-grade meme coin sniping bot for BSC mainnet. Written in Go, with a Next.js 14 dashboard, PostgreSQL, and Redis.

---

## Architecture

```
PancakeSwap Factory ──WebSocket──▶ Listener
                                       │
                              Redis stream:new_pairs
                                       │
                                   Filter Engine
                              (liquidity, age, holders,
                               honeypot simulation)
                                       │
                            Redis stream:approved_tokens
                                       │
                                   Executor (Buy)
                              PancakeSwap Router swap
                                       │
                                Position Monitor
                              (200ms price polling,
                               TP1, trailing stop,
                               time-based exit)
                                       │
                                   Executor (Sell)
```

---

## Quick Start

### 1. Prerequisites

- Docker + Docker Compose v2
- A funded BSC wallet (≥0.01 BNB recommended for gas)

### 2. Configure

```bash
cp .env.example .env
# Edit .env — fill in your RPC URLs and private key
```

**Important `.env` fields:**

| Variable | Description |
|---|---|
| `BSC_RPC_WS` | NodeReal / Ankr / BNB Chain WebSocket endpoint |
| `BSC_RPC_HTTP` | NodeReal / Ankr HTTP endpoint |
| `PRIVATE_KEY` | Your wallet private key (starts with 0x) |
| `LIVE_TRADING_ENABLED` | `false` = simulation, `true` = real trades |
| `BUY_AMOUNT_BNB` | BNB spent per trade (default 0.0005) |

### 3. Build & Run

```bash
cd bsc-sniper
docker compose build
docker compose up -d
```

### 4. Verify

```bash
# Health check
curl http://localhost:8080/api/health

# Start the bot
curl -X POST http://localhost:8080/api/bot/start

# Watch logs
docker compose logs -f backend | grep -E "PairCreated|Buy executed|Sell executed"
```

### 5. Dashboard

Open http://localhost:3000 in your browser.

---

## Safety Filters

Each new pair goes through these checks in parallel (2s timeout):

| Check | Default | Purpose |
|---|---|---|
| Min liquidity | $1,000 USD | Avoid illiquid scams |
| Max age | 300 sec | Only fresh pairs |
| Min holders | 25 | Early-stage community |
| Top-10 concentration | ≤35% | Whale rug protection |
| Honeypot simulation | — | Buy & sell simulation via router |

---

## Exit Strategy

1. **Take-Profit 1** at +100% → sell 50% of position
2. After TP1, activate **25% trailing stop** from ATH
3. If trailing stop triggered → sell remaining 100%
4. **Time-based exit**: position open >2h and P&L <20% → market sell 100%

---

## API Reference

| Method | Path | Description |
|---|---|---|
| GET | `/api/health` | Health check & stats |
| POST | `/api/bot/start` | Start the sniping bot |
| POST | `/api/bot/stop` | Stop the bot |
| GET | `/api/bot/status` | Detailed bot status |
| GET | `/api/tokens` | Recent tokens discovered |
| GET | `/api/trades` | Trade history |
| GET | `/api/positions` | Current & past positions |
| GET | `/api/config` | Current configuration |
| PUT | `/api/config` | Update configuration live |
| WS | `/api/ws` | Real-time event stream |

---

## MEV Protection

Set `BLOXROUTE_URL` to your BloxRoute private relay endpoint to submit transactions privately and avoid front-running. When empty, transactions go through the public BSC mempool.

---

## Security Notes

- **Never commit `.env`** — it is in `.gitignore`.
- The private key is never logged or exposed via API.
- All external calls use context timeouts to prevent hangs.
- Run with `LIVE_TRADING_ENABLED=false` first to verify filters are working correctly.

---

## Stack

| Component | Technology |
|---|---|
| Backend | Go 1.21, go-ethereum, Gin |
| Blockchain | PancakeSwap V2, BSC mainnet |
| Database | PostgreSQL 16 |
| Cache / Bus | Redis 7 (streams + pub/sub) |
| Frontend | Next.js 14, TailwindCSS, shadcn/ui, Recharts |
| Deployment | Docker Compose |
