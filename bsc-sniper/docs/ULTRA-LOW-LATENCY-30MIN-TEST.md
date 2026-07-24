# Ultra-Low-Latency 30-Minute Live BSC Mainnet Test

**Date:** 2026-07-24 UTC  
**Runtime:** 30 minutes (23:39:02 → 00:09:30 UTC)  
**Filters:** efficiency ≥ 0.95, liquidity ≥ $12,000, WBNB quote only, buy size 0.0005 BNB  
**Branch:** main  
**Commit context:** mempool V2 listener, 1.5× buy gas / 2.0× retry gas, post-buy simulation + emergency dump, no liquidity retry.

## Summary

| Metric | Value |
|---|---|
| Pairs seen | 22 |
| Pairs passing filter | 5 |
| Buy attempts | 5 |
| Successful sells | 3 |
| Unsold positions | 2 |
| Realized P&L (BNB) | **-0.00000555** |
| Unrealized at risk (BNB) | 0.00100 |
| Net cost exposure (BNB) | 0.00250 |
| Mempool V2 pairs detected | 0 (subscription stayed active; no `mempool_v2` events were observed) |
| V2 log pairs | 20+ |
| V3 log pairs | 1 (USDC quote, rejected) |
| StableSwap log pairs | 0 |

## Key code changes

1. **Mempool V2 listener** (`backend/internal/listener/pancake.go`): added `eth_subscribe("newPendingTransactions")` path, decodes PancakeSwap V2 `createPair` inputs, computes the deterministic pair address via CREATE2, and publishes `NewPairEvent` with source `mempool_v2` without waiting for a block. V2/V3/StableSwap log subscriptions remain as fallback.
2. **Liquidity check** (`backend/internal/filter/engine.go`): `getLiquidityUSD` now checks exactly once and rejects immediately; the 3-second retry is removed.
3. **Buy execution** (`backend/internal/executor/buy.go`): buy gas = 1.5×, on failure one retry at 2.0× with no sleep; new positions are `bought`; added `postBuyVerification` and `tryEmergencyDump` (50% slippage, 2.0× gas); `ExecuteSellCustom` added for emergency sells.
4. **Position monitor** (`backend/internal/monitor/position.go`): now loads `bought` and `partial` positions instead of `open`.
5. **Stop/unsold board** (`backend/cmd/main.go`): `Stop()` prints a console table of `bought`/`unsellable` positions and writes contract addresses to `unsold_tokens.txt`.
6. **DB schema** (`backend/internal/db/schema.sql`): widened `positions.status` check constraint to allow `bought` and `unsellable`.

## Positions

| Symbol | Token | Status | Cost BNB | Realized P&L BNB | Notes |
|---|---|---|---|---|---|
| COSA | `0xfD2d10ADE95499edf7eab373259A05e0bc45614e` | **bought** | 0.00050 | 0.00000 | Post-buy simulation efficiency 0.9950; later sell attempts reverted; force sell reverted on stop. |
| CZ蝴蝶 | `0x50527829E0C491eb104bae3e0feDF95265f884C7` | closed | 0.00050 | -0.00000055 | Sold via breakeven stop-loss. |
| USAR | `0x2F6cD2925204c05a809523f08e417dE10Ba47faC` | **bought** | 0.00050 | 0.00000 | Post-buy simulation efficiency 0.9950; subsequent sells reverted; force sell reverted on stop. |
| 王纯 | `0x4a9E77D8Fa1D28C0e14f1418950E30F45399379C` | closed | 0.00050 | -0.00000250 | Sold via breakeven stop-loss. |
| WDC | `0x43b243aec6D9300591CB2AA747a97f13B9cD8e31` | closed | 0.00050 | -0.00000250 | Sold via breakeven stop-loss. |

## Unsold token board (printed on stop)

```
═══════════ UNSOLD TOKEN BOARD ═══════════
Symbol       | Contract                                   | Buy Tx                                                             | Amount                   | Status      
---------------------------------------------------------------------------------------------------------------------------------------------------------------------
USAR         | 0x2F6cD2925204c05a809523f08e417dE10Ba47faC | 0xada6828be0dc6402ed42606cbfc96d6c0308ac8bc70b7ad6006d86468ef22ee0 | 19949602005439           | bought      
COSA         | 0xfD2d10ADE95499edf7eab373259A05e0bc45614e | 0xb10fd553170284fc7fc0454c7ae292a5d92676d7518778096aef33fd105b29b9 | 3324889447425            | bought      
═══════════════════════════════════════════
```

`unsold_tokens.txt` contents:

```
0x2F6cD2925204c05a809523f08e417dE10Ba47faC
0xfD2d10ADE95499edf7eab373259A05e0bc45614e
```

## Observations

- **No mempool pairs detected.** The BSC WebSocket provider accepted the `newPendingTransactions` subscription and the listener stayed active, but zero V2 `createPair` calls were observed during the 30-minute window. All pairs were received via the fallback V2 log subscription. This is consistent with most public BSC nodes not exposing a true mempool or dropping pending transactions, but it means the deterministic CREATE2 pair-address computation path was not exercised in production during this run.
- **Post-buy verification passed for every buy.** All five buys had a simulated sell efficiency of ~0.9950, so no emergency dumps were triggered at buy time. This suggests the tokens were not honeypots at the moment of purchase (the pre-buy filter also rejected several low-efficiency / low-liquidity candidates).
- **Later sell attempts failed for the two unsold tokens.** COSA and USAR repeatedly hit reverted sell transactions when the monitor tried to exit them via breakeven stop-loss or the final force sell. This indicates a transfer/sell tax or contract restriction that activates after purchase, even though the post-buy simulation succeeded. These positions are effectively unsellable at current slippage/gas.
- **No new V3/StableSwap opportunities.** Only one V3 pool appeared (USDC quote, rejected by WBNB-only rule), and no StableSwap pairs.

## P&L calculation

- **Realized:** 3 sells received `0.00049945 + 0.00049750 + 0.00049750 = 0.00149445` BNB.
- **Cost of sold positions:** 3 × 0.00050 = 0.00150 BNB.
- **Realized loss:** `0.00149445 - 0.00150 = -0.00000555` BNB.
- **Unrealized / at-risk:** 2 × 0.00050 = 0.00100 BNB (COSA, USAR).
- **Total deployed:** 0.00250 BNB.

## Latency notes

Because the mempool listener produced zero actionable events during the run, measured latency is based on the log-subscription path (pair event published once the `PairCreated` log is included in a block). The buys were submitted within seconds of the pair event, and the post-buy sell simulation ran immediately after the buy receipt confirmed. Removing the 3-second liquidity retry eliminated that source of delay entirely.

## Follow-up items

1. **Mempool fallback:** verify whether the BSC provider truly supports `newPendingTransactions` or whether an alternative streaming endpoint (e.g., BloXroute, Eden) is needed. If public RPCs do not support it, consider a more resilient fallback that listens to pending transactions via a local node or paid mempool stream.
2. **Post-buy detection of sell-blocking contracts:** the two unsold tokens passed simulation but could not be sold later. Investigate whether a deeper simulation (e.g., simulating a second transfer after a small delay, or checking for dynamic blacklists) can catch these before buying.
3. **Force-sell error clarity:** the log message `send sell failed after retry: %!s(<nil>)` has been fixed in `backend/internal/executor/buy.go` to distinguish between send errors and revert/no-receipt failures.

## Files changed

- `backend/internal/db/schema.sql`
- `backend/internal/db/db.go`
- `backend/internal/listener/pancake.go`
- `backend/internal/filter/engine.go`
- `backend/internal/executor/buy.go`
- `backend/internal/monitor/position.go`
- `backend/cmd/main.go`
- `unsold_tokens.txt`
- `docs/ULTRA-LOW-LATENCY-30MIN-TEST.md`
