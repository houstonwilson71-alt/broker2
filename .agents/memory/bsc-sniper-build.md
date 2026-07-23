---
name: BSC Sniper build decisions
description: Key technical decisions, version pins, and quirks for the bsc-sniper project
---

## Go module / version constraints
- pgx/v5 pinned to **v5.6.0** — v5.5.4 has invalid pseudo-version; v5.10.0 requires Go 1.25
- Go 1.21 (golang:1.21-alpine builder)
- `GOPROXY=https://proxy.golang.org,direct GONOSUMDB='*'` required for `docker compose build` — Replit firewall blocks some transitive go-ethereum deps

## go-ethereum API notes
- Use `ethereum.CallMsg` NOT `bind.CallMsg` — the bind package does not export CallMsg
- `types.NewEIP155Signer(chainID)` for BSC mainnet (chain ID 56)

## Docker / Compose constraints
- Docker healthchecks (`docker exec`-based) fail in Replit's sandboxed OCI (setns restriction)
- Use `condition: service_started` NOT `condition: service_healthy`
- Backend makes 3 rapid restart attempts before Postgres finishes init — this is normal, not a bug

## Architecture decisions (hardened build)
- Single WebSocket subscription to all 3 factories (V2+V3+StableSwap) with combined FilterQuery
- 4 filter workers + 4 executor workers — each with unique consumer ID (`consumer:filter:N`)
- Ring-buffer dedup: 10k slot LRU via map + array, O(1) lookup
- Rate limiter: token-bucket 100 req/s shared via channel
- Circuit breaker: 3 consecutive tx failures → 2-min buy pause
- Gas price: suggestedGasPrice × 1.5 for aggressive inclusion
- Buy function: `swapExactETHForTokensSupportingFeeOnTransferTokens` for all buys (handles fee tokens)
- Monitor graceful shutdown: `persistAllPositions()` called before exit
- `min()` built-in conflict in Go 1.21: rename custom duration helper to `minDur()`
- Local type `tokenInfoVal` inside `processEvent` shadowed package-level type → runtime panic; fix by removing local definition

## Factory addresses (BSC mainnet)
- V2: 0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73 (PairCreated — active but rarely WBNB pairs)
- V3: 0x0BFbCF9fa4f9C56B0F40a671Ad40E0805A091865 (PoolCreated — very low activity)
- StableSwap: 0x25a55f9f2279A54951133D503490342b50E5cd15 (NewStableSwapPair)
- WBNB: 0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095b (lowercase: bb4cdb...095b ends in b)

## Market observations (2026-07-23)
- PancakeSwap V2 WBNB-paired new listings: essentially zero (confirmed over 5000-block window)
- New tokens predominantly use USDT/BUSD/USDC as base or list on ApeSwap/BiSwap/BabySwap
- To get actual trade volume, must expand base-token filter beyond WBNB only

## GitHub
- Repo: https://github.com/houstonwilson71-alt/broker2 (private)
- Push requires --force (remote had existing commits)

## Next.js version
- Pin to 14.2.35 (14.2.3 blocked by Replit security policy)
