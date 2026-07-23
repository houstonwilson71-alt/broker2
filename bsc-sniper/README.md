# BSC Meme-Coin Sniper — Hardened Multi-Version Edition

Production-grade BSC mainnet sniper bot. Listens to **PancakeSwap V2 + V3 + StableSwap** factory events, applies rigorous safety filters, and executes buys with surgical precision.

## Architecture

```
WebSocket Listener (V2+V3+Stable) ──► Redis stream:new_pairs
    │ ring-buffer dedup                        │
    │ exponential-backoff reconnect     4× Filter Workers
    │                                   ├─ Liquidity check    ┐
    │                                   ├─ Token age          ├─ parallel 2.5s deadline
    │                                   ├─ Honeypot sim       │
    │                                   └─ Holder check       ┘
    │                                          │
    │                                   Redis stream:approved_tokens
    │                                          │
    │                                   4× Executor Workers
    │                                   ├─ Gas price × 1.5
    │                                   ├─ Dual-submit (RPC + BloxRoute)
    │                                   ├─ Circuit breaker (3-fail → 2-min pause)
    │                                   └─ On-chain receipt wait
    │                                          │
    │                                   Position Monitor (200ms tick)
    │                                   ├─ TP1: +100% → sell 50%
    │                                   ├─ Trailing stop: -25% from ATH → sell 100%
    │                                   └─ Time exit: >2h, <20% gain → sell 100%
    │
PostgreSQL 16 ◄── All events persisted
Redis 7       ◄── Streams + pub/sub
Next.js 14    ──► Dashboard (port 3000)
Gin REST API  ──► Backend (port 8080)
```

## Quick Start

```bash
cp .env.example .env
# Edit .env with your credentials

docker compose build
docker compose up -d

# Start the bot
curl -X POST http://localhost:8080/api/bot/start

# Dashboard
open http://localhost:3000
```

## Factory Coverage

| Exchange | Factory | Event | Status |
|---|---|---|---|
| PancakeSwap V2 | `0xcA143Ce32...` | `PairCreated` | ✅ Subscribed |
| PancakeSwap V3 | `0x0BFbCF9fa...` | `PoolCreated` | ✅ Subscribed |
| PancakeSwap StableSwap | `0x25a55f9f...` | `NewStableSwapPair` | ✅ Subscribed |

All three subscribed in a **single WebSocket connection** with a combined filter query.

## Safety Filters

All checks run in parallel with a strict **2.5-second total deadline**:

1. **Liquidity** — `getReserves()` + BNB price (CoinGecko, 30s cache, on-chain fallback)
2. **Age** — reject if pair event > `MAX_AGE_SEC` seconds old
3. **Honeypot simulation** — `getAmountsOut` buy then sell; reject if sell tax > 50%
4. **Holder count** — BSCscan API (best-effort, non-blocking)

## Hardening Features

| Feature | Implementation |
|---|---|
| Ring-buffer dedup | 10,000-slot LRU, O(1) lookup |
| WS auto-reconnect | Exponential backoff 1s→30s, resets on stable connection |
| Rate limiter | Token-bucket 100 req/s shared across all workers |
| Circuit breaker | 3 consecutive tx fails → 2-minute buy pause |
| Panic recovery | Monitor tick recovers from any panic, logs stack |
| Graceful shutdown | Open positions flushed to DB before exit |
| Emergency stop | `POST /api/bot/emergency-stop` → sells all positions |

## API Reference

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/health` | Health check + state |
| `POST` | `/api/bot/start` | Start the bot |
| `POST` | `/api/bot/stop` | Graceful stop |
| `POST` | `/api/bot/emergency-stop` | Sell all + stop |
| `GET` | `/api/bot/status` | Status, counters, state |
| `GET` | `/api/trades` | Trade history |
| `GET` | `/api/positions` | Open/closed positions |
| `GET` | `/api/tokens` | Detected tokens |
| `GET` | `/api/config` | Current config |
| `PUT` | `/api/config` | Update config live |
| `GET` | `/api/ws` | WebSocket event stream |

## Exit Strategy

```
Buy → Monitor (200ms)
  If price +100%:         sell 50% (TP1)
  If ATH drawdown -25%:  sell 100% (trailing stop, after TP1)
  If open >2h, gain <20%: sell 100% (time exit)
```

## Environment Variables

See `.env.example` for all options.

## Notes

- **Private key** — never logged, stored only in `.env` (gitignored)
- **WBNB filter** — strict exact address match; near-WBNB scam tokens are rejected
- **V2 vs V3** — PancakeSwap V2 new-pair activity has declined significantly; V3 and StableSwap coverage ensures comprehensive detection
- **BloxRoute** — set `BLOXROUTE_URL` to enable private mempool submission
