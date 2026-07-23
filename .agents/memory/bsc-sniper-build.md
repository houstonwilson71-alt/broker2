---
name: BSC Sniper build decisions
description: Key environment quirks, confirmed fixes, and test findings for the bsc-sniper project
---

## Environment

- pgx/v5 v5.6.0 required (later versions break API)
- Go build uses `bind.CallMsg` → must use `ethereum.CallMsg` instead
- Go module proxy must be bypassed: `GONOSUMDB=* GOPROXY=off` when building offline
- `docker compose exec -T` fails with OCI errors in Replit; use `docker exec <container>` or the HTTP API instead

## WBNB Address

**Always use `0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c`** (ends in `bc095c`, NOT `bc095b`).
The typo `bc095b` was present in 6 files; corrected 2026-07-23. The wrong address silently routes everything through a dead address.

## V3 Pool Liquidity

V3 pools do NOT have `getReserves()`. Use `balanceOf(poolAddress)` on the quote token ERC-20 to measure liquidity. This is also valid for V2. Always use `balanceOf` — it's universal.

## Post-PairCreated Liquidity Timing

PairCreated fires 1 block before liquidity is added. Use a 3×2s retry loop inside `getLiquidityUSD`. Without retries, ~30% of real pairs are incorrectly rejected as zero-liquidity.

**Why:** The factory emits the event when the pair contract is deployed. The LP add is a separate transaction by the creator, usually in the next block (~3 s on BSC).

## PriceBNB DB Insert

The `trades` table has a `NOT NULL` `price_bnb` NUMERIC column. On the buy side, set `PriceBNB: "0"` as a placeholder — never leave it as an empty string.

## Multi-Quote 2-Hop Routing

USDT/BUSD/USDC/ETH/CAKE pairs route as: stable→WBNB→memeToken (2-hop). path_hops=1 means WBNB direct. path_hops=2 means stable via WBNB. Both confirmed working live on mainnet (2026-07-23).

## Live Test Findings (2026-07-23, 60 min)

- **Revert rate ~32%**: Most reverts are honeypots. Pre-buy `eth_call` simulation (buy+sell) recommended before shipping.
- **Wallet depletion**: Test wallet drained after ~22 buys (0.0005 BNB × 22 + gas). Production needs ≥ 0.05 BNB.
- **Symbol collisions normal**: Multiple tokens can share the same symbol (GANA, QQQB seen twice each). Bot correctly buys by address, not symbol.
- **BUSD/ETH/CAKE**: Rarely used as quote tokens for new meme launches. Not seen in 60-min window — code handles correctly when they appear.
- **status updater gap**: Reverted on-chain txns may stay `pending` in DB for up to ~30s; status-updater should be tightened.
