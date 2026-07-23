# 30‑Minute Live Mainnet Test — V3/StableSwap Listener + 3‑Second Liquidity Retry

**Date:** 2026‑07‑23  
**Test window:** 22:33:34 UTC → 23:03:34 UTC (30 minutes)  
**Wallet:** `0xA2591F919f18Ba5b6f8A00f872dB07fce968Ef84`  
**Network:** BSC mainnet  
**Branch/commit:** `main` (to be committed after this report)

---

## 1. What changed

| Setting | This test | Previous strict WBNB‑only test |
|---|---|---|
| Efficiency threshold | **0.95** | 0.95 |
| Liquidity floor | **$12,000** | $12,000 |
| Quote‑token guard | **WBNB only** | WBNB only |
| Buy size | **0.0005 BNB** | 0.0005 BNB |
| Listener | **V2 + V3 (all fee tiers) + StableSwap** | V2 only |
| Liquidity retry | **Immediate check → 3 s wait → recheck** | Immediate check with 2 s retries |
| TP / SL | Partial TP (50% at +200%, trailing stop) | Same |
| Slippage / gas | 15% buy, 5% sell, 1.5× gas boost on sells | Same |

### Code changes
- `docker-compose.yml`: `MIN_LIQUIDITY_USD` restored to **12000**.
- `backend/internal/filter/engine.go`:
  - `eliteLiquidityFloorUSD` restored to **12000.0**.
  - `eliteMinRoundTripRatio` restored to **0.95**.
  - `getLiquidityUSD` now checks once, waits **3 seconds**, and checks again if the first reading is below the $12 k floor; rejects if it is still below.
- `backend/internal/listener/pancake.go`:
  - V3 `PoolCreated` events are explicitly accepted for **all fee tiers** (0.05%, 0.25%, 1%).
  - Added `FeeTier` to `NewPairEvent` and logs for transparency.

---

## 2. Results summary

| Metric | This test | Previous strict WBNB‑only test (20 min) |
|---|---|---|
| Pairs seen | 31 | 3 |
| Pairs passed filter | 9 | 3 |
| Confirmed buys in window | **8** | 3 |
| Confirmed sells in window | 7 | 3 |
| Failed buy in window | 1 (replacement underpriced, no gas loss) | 0 |
| Unsold leftovers | **2 (WCDOGE, SPY)** | 0 |
| Net trade‑based P&L (window) | **-0.00237253 BNB** | -0.000008 BNB |
| Approx. USD (@ $566/BNB) | **-$1.34** | ~-$0.0045 |

### P&L detail (trade‑based, 30‑min window)
- Total BNB spent on confirmed buys: **0.00400000 BNB** (8 × 0.0005)
- Total BNB recovered from confirmed sells: **0.00162747 BNB**
- **Net P&L: -0.00237253 BNB**

The two unsold leftovers (WCDOGE, SPY) are still in the wallet but cannot be sold through PancakeSwap V2 — their sell transactions revert even with 50% slippage and 1.5× gas, indicating sell‑restricted contracts. If valued at zero, the realized/unrealized loss is the same -0.00237 BNB; if they remain unsellable, they are a full 0.0010 BNB additional loss on top of the 0.00237 BNB already realized by the seven sells.

### Note on an accidental extra buy
During the post‑window force‑sell cleanup, the bot briefly restarted and bought one additional WBNB pair (`0x96A8...ef1AbD`, token amount 33,248,894,474,258). This transaction was confirmed on‑chain (tx `0xc4d5...9ade3`) but the bot was interrupted before recording the receipt. It is **outside the 30‑minute test window**. Including it in the P&L gives -0.00287253 BNB; the report focuses on the 30‑minute window and treats this as a cleanup artifact.

---

## 3. Trade ledger (30‑minute window)

| # | Token | Side | BNB | Status | Tx hash |
|---|-------|------|-----|--------|---------|
| 1 | WCDOGE | buy | 0.00050 | confirmed | `0x06f2...9e684` |
| 2 | ZYLO | buy | 0.00050 | confirmed | `0x4f6e...c51878` |
| 3 | ZYLO | sell | 0.00049750 | confirmed | `0x2380...84ce37` |
| 4 | UniHGPT | buy | 0.00050 | confirmed | `0xb8f3...40b3a` |
| 5 | UniHGPT | sell | 0.00000300 | confirmed | `0x0dce...c9c7e` |
| 6 | UniHGPT | sell | 0.00000012 | confirmed | `0x4cfa...3d6e00` |
| 7 | UniFCHAIN | buy | 0.00050 | confirmed | `0x0fc1...28dff6` |
| 8 | UniFCHAIN | sell | 0.00049750 | confirmed | `0x6acd...6b0e30` |
| 9 | UniCATFI | buy | 0.00050 | confirmed | `0x0a6c...113d9f` |
| 10 | 天梯 | buy | 0.00050 | failed | replacement underpriced |
| 11 | 天梯 | buy | 0.00050 | confirmed | `0x2450...ef63fa` |
| 12 | UniCATFI | sell | 0.00000280 | confirmed | `0x5b03...530540` |
| 13 | SKY🔶BNB | buy | 0.00050 | confirmed | `0xb23e...559bd9` |
| 14 | SKY🔶BNB | sell | 0.00049750 | confirmed | `0x9746...680877` |
| 15 | SPY | buy | 0.00050 | confirmed | `0xd1ba...194a37` |
| 16 | 天梯 | sell | 0.00012905 | confirmed | `0x9935...bb0fb8` |

**Unsold leftovers:**
- WCDOGE (`0xa453...33B5e`) — multiple sell attempts revert.
- SPY (`0x17B0...A569`) — multiple sell attempts revert.

---

## 4. Filter observations

### Liquidity retry behavior
The new 3‑second retry rejected several pairs whose liquidity was just below $12 k at the second check. Examples from the log:

| Token | First check | Second check (after 3 s) | Outcome |
|-------|-------------|--------------------------|---------|
| `0xA155...db8d97` | <$12 k | $2,834 | rejected |
| `0x9006...cb3Ed6` | <$12 k | $8,850 | rejected |
| `0xF865...862305` | <$12 k | $11,691 | rejected |
| `0x846e...1dF0a` | <$12 k | $9,196 | rejected |

This confirms the retry is functioning, but it also means we are missing pairs that deploy with liquidity just under $12 k and never cross the threshold.

### V3 / StableSwap
The listener is subscribed to V2, V3, and StableSwap factories. During the 30‑minute window, **all purchased tokens came from PancakeSwap V2 WBNB pairs**. No V3 or StableSwap WBNB pair passed the filter (several V2 USDT pairs were rejected by the WBNB‑only guard before reaching the V3/StableSwap relevance).

### WBNB guard
USDT/BUSD/USDC/ETH/CAKE pairs continued to be rejected immediately. No stablecoin tax trap occurred in this run.

### Honeypots / sell tax
The most damaging token was **UniHGPT** (trade #4): bought 0.0005 BNB, partial TP sold 0.000003 BNB, trailing/force sell recovered only 0.00000012 BNB. Combined recovery: **0.00000312 BNB** on a 0.0005 BNB cost — a 99.4% effective loss. This matches the previous test finding: pre‑buy static simulation does not predict the live sell tax that hits after the first block.

---

## 5. Comparison to previous strict WBNB‑only test

| | Previous strict WBNB‑only (20 min) | This test (30 min) |
|---|---|---|
| Trades | 3 | 8 buys + 7 sells |
| Net P&L | **-0.000008 BNB** (near flat) | **-0.00237253 BNB** |
| Loss per trade | ~-0.0000027 BNB | ~-0.0002966 BNB per buy |
| Key difference | Tight filter + $12 k floor caught only 3 clean pairs | Expanded listener + 3 s retry still caught 8, but 2 were unsellable honeypots and several sold at near‑total loss |

**Interpretation:** widening the listener surface (V3 + StableSwap) did not materially increase the number of passing pairs in this window, but the **liquidity retry allowed more marginal WBNB V2 pairs through**. Those marginal pairs had higher honeypot/sell‑tax incidence, turning the near‑flat previous result into a ~-0.0024 BNB loss. The 3‑second retry is working as designed, but it is catching tokens that barely clear the $12 k floor and are more likely to be tax traps.

---

## 6. Recommendations for next test

1. **Keep the liquidity retry, but require a higher minimum simulated sell output.** For example, reject if the pre‑buy simulation returns less than 0.00045 BNB on a 0.0005 BNB buy. This directly targets the “near‑total loss” tokens like UniHGPT.
2. **Tighten the efficiency threshold further** (e.g. 0.97) or add a second simulation after the 3‑second liquidity wait to catch tokens whose tax activates after deployment.
3. **Skip the liquidity retry for non‑WBNB pairs** — already rejected by the WBNB guard, so no change needed.
4. **Consider a minimum absolute BNB recovery per sell** in the monitor: if a breakeven SL would recover less than a threshold (e.g. 0.00045 BNB), skip the sell to avoid wasting gas on already‑taxed tokens. (This is more of a gas‑saving measure than a P&L improvement.)
5. **Investigate the sell‑revert bug:** the executor currently returns a nil error message (`%!s(<nil>)`) on revert. Better error logging would clarify whether the reverts are slippage‑related or contract‑level.

---

## 7. Files changed

- `docker-compose.yml`
- `backend/internal/filter/engine.go`
- `backend/internal/listener/pancake.go`
- `docs/ANALYTICAL-FILTER-TEST-30MIN-V3-RETRY.md` (this report)
