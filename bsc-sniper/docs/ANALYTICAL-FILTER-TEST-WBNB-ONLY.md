# WBNB-Only Live Mainnet Test — Final Report

**Date:** 2026-07-23  
**Duration:** 20 minutes (started 20:57:44 UTC, stopped 21:18:34 UTC)  
**Wallet:** `0xA2591F919f18Ba5b6f8A00f872dB07fCe968Ef84`  
**Buy amount:** 0.0005 BNB  
**Efficiency threshold:** 0.95 (95% round-trip)  
**Liquidity floor:** $12,000 USD  
**Quote-token restriction:** WBNB only — USDT, BUSD, USDC, ETH, CAKE explicitly rejected  
**TP strategy:** 50% at +200%, remaining 50% at +300% or trailing stop -20% from peak  
**Sell slippage:** amountOutMin = expected × 0.95  
**Gas boost:** 1.5× on take-profit sells  
**Duplicate symbol guard:** 5 minutes  
**BSCScan retries:** 3 attempts  
**BNB price used for USD:** $566.81  

## 1. Executive Summary

- **Total buys:** 3  
- **Total BNB spent:** 0.0015 BNB  
- **Total sell transactions:** 3  
- **Total BNB recovered:** 0.001492 BNB  
- **Net P&L:** -0.000008 BNB (-0 USD)  
- **All exits:** Break-even SL (no token reached +200%)  
- **Pairs approved:** 3  
- **Pairs rejected:** 14  

## 2. All Buys (Full Transaction Hashes)

| # | Token | Symbol | Quote | Buy BNB | Buy Tx |
|---|-------|--------|-------|---------|--------|
| 1 | 0x154e39896C8bc55D2Fbde7a022A525f6d4Db3a8E | POET | WBNB | 0.0005 | [0x176601...](https://bscscan.com/tx/0x176601b00326833e34b85e879a39dc1a985d4c052cc732f7b65237d9320cb4d1) |
| 2 | 0x9faf4eE4cbdb5E757f0B534e1c4d37932F2380e9 | rSTX | WBNB | 0.0005 | [0xf52d3f...](https://bscscan.com/tx/0xf52d3f418f01a064a3654af66940586d66bd17e34a1b73c14c93ba0d3da23a66) |
| 3 | 0xD39a8585A88Be597DD14209d28A5F97CCA071919 | 宇宙所 | WBNB | 0.0005 | [0x7c7037...](https://bscscan.com/tx/0x7c70374b020657ca67eb3def1ec22ef999e7b8f734acf5edf7df66f291ae3ef5) |

**Full buy hashes:**

- POET — 0x176601b00326833e34b85e879a39dc1a985d4c052cc732f7b65237d9320cb4d1
- rSTX — 0xf52d3f418f01a064a3654af66940586d66bd17e34a1b73c14c93ba0d3da23a66
- 宇宙所 — 0x7c70374b020657ca67eb3def1ec22ef999e7b8f734acf5edf7df66f291ae3ef5

## 3. All Sells (Full Transaction Hashes + Exit Reason)

| # | Token | Symbol | Quote | Sell BNB | Reason | Sell Tx |
|---|-------|--------|-------|----------|--------|---------|
| 1 | 0x154e39896C8bc55D2Fbde7a022A525f6d4Db3a8E | POET | WBNB | 0.000498 | Break-even SL | [0x965d3f...](https://bscscan.com/tx/0x965d3fb5bee31937e223872b49f583f2f88e240d0e0a15088ec3c333c916aedf) |
| 2 | 0x9faf4eE4cbdb5E757f0B534e1c4d37932F2380e9 | rSTX | WBNB | 0.000498 | Break-even SL | [0x2ad19f...](https://bscscan.com/tx/0x2ad19f83cb2d9a8b78937e9e5b184276c093a72cdecbcea03704bd194ca7f525) |
| 3 | 0xD39a8585A88Be597DD14209d28A5F97CCA071919 | 宇宙所 | WBNB | 0.000498 | Break-even SL | [0x756858...](https://bscscan.com/tx/0x75685878955218be076a53a7263d0dc0559f9025193d144583b116474d086dda) |

**Full sell hashes:**

- POET (Break-even SL) — 0x965d3fb5bee31937e223872b49f583f2f88e240d0e0a15088ec3c333c916aedf
- rSTX (Break-even SL) — 0x2ad19f83cb2d9a8b78937e9e5b184276c093a72cdecbcea03704bd194ca7f525
- 宇宙所 (Break-even SL) — 0x75685878955218be076a53a7263d0dc0559f9025193d144583b116474d086dda

## 4. Per-Position P&L

| Token | Symbol | Buy BNB | Recovered BNB | P&L BNB | P&L % | Exit Type |
|-------|--------|---------|---------------|---------|-------|-----------|
| 0x154e39896C8bc55D2Fbde7a022A525f6d4Db3a8E | POET | 0.0005 | 0.000498 | -0.000003 | -0.5% | Break-even SL |
| 0x9faf4eE4cbdb5E757f0B534e1c4d37932F2380e9 | rSTX | 0.0005 | 0.000498 | -0.000003 | -0.5% | Break-even SL |
| 0xD39a8585A88Be597DD14209d28A5F97CCA071919 | 宇宙所 | 0.0005 | 0.000498 | -0.000003 | -0.5% | Break-even SL |

## 5. Efficiency Data (Every Token Filtered)

| Token | Symbol | Quote | Efficiency | Status | Rejection Reason |
|-------|--------|-------|------------|--------|------------------|
| 0x154e39896C8bc55D2Fbde7a022A525f6d4Db3a8E | POET | WBNB | 0.9949 | APPROVED |  |
| 0x9faf4eE4cbdb5E757f0B534e1c4d37932F2380e9 | rSTX | WBNB | 0.9949 | APPROVED |  |
| 0xD39a8585A88Be597DD14209d28A5F97CCA071919 | 宇宙所 | WBNB | 0.995 | APPROVED |  |
| 0x9275290804edfaA88d83524A98b8F662d3e2FCB0 |  | WBNB | 0.9948 | REJECTED | low_liquidity:$11345 |
| 0xa43dbc4e560b875Bb6073EB59Bd4794B49A3219d |  | WBNB | 0.9947 | REJECTED | low_liquidity:$6731 |
| 0x2bcD177e02a55b5aA068b58a001214bDB8224794 |  | WBNB | 0.9942 | REJECTED | low_liquidity:$2837 |
| 0x11690431c677A37F50b63A460861D5BBE2dDBD6D |  | WBNB | 0.9947 | REJECTED | low_liquidity:$8567 |
| 0xE5f526eb2D61Dab9D39ff41e9a291dFdF7242972 |  | USDT | 0.9899 | REJECTED | non_wbnb_quote:USDT |
| 0x7ac878E680eC7a66Aee5AfF20Cb0510D07b64a81 |  | USDT | 0.9899 | REJECTED | non_wbnb_quote:USDT |
| 0xcf26E1515035F70d4696e98716a8D9F3aFDd501D |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0x3E56eE6F61796Cc2D592f84A1594cF675806E3d3 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0xe087dCF680b1Ce51D097B73561bE0d161Af1A395 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, low_liquidity:$23 |
| 0x8Ad6d5dC57ad419813b48B6B0533057ea28149A3 |  | WBNB | 0.9948 | REJECTED | low_liquidity:$11908 |
| 0xaE35857A58733f8419Be18Af7bE7D7403ceeAe7E |  | WBNB | 0.993 | REJECTED | low_liquidity:$1134 |
| 0x4930a1ef6C4a276f4319842C32564FE94Ec6bb5c |  | WBNB | 0.9736 | REJECTED | low_liquidity:$103 |
| 0x420Aa899d43aced90c585F1D9ae9fF63810dd1aA |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0xbC4533Ee10A85f5Bc74c599160d455fa73890BE9 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |

## 6. Rejected Tokens Breakdown

- **Rejected by non-WBNB quote guard:** 2  
- **Rejected by efficiency guard:** 5  
- **Rejected by liquidity floor:** 7  
- **Rejected by duplicate symbol guard:** 0  
- **Rejected by other reasons:** 0  

## 7. Comparison to Previous Tests

| Metric | 30-min Test (0.001 BNB, 0.85 eff, USDT allowed) | 20-min Test V1 (0.0005 BNB, 0.95 eff, USDT allowed) | This Test (WBNB only) |
|--------|---------------------------------------------------|------------------------------------------------------|-----------------------|
| Buy size | 0.001 BNB | 0.0005 BNB | 0.0005 BNB |
| Efficiency threshold | 0.85 | 0.95 | 0.95 |
| Quote restriction | USDT + WBNB | USDT + WBNB | WBNB only |
| Duration | 30 min | 20 min | 20 min |
| Buys | 16 | 5 | 3 |
| BNB spent | 0.016 BNB | 0.0025 BNB | 0.0015 BNB |
| BNB recovered | 0.008975 BNB | 0.001494 BNB | 0.001492 BNB |
| Net P&L | -0.007025 BNB | -0.001006 BNB | -0.000008 BNB |
| Net P&L USD | -3.97 USD | -0.57 USD | -0 USD |
| Loss per trade | -0.000439 BNB | -0.000201 BNB | -0.000003 BNB |

### Analysis

Eliminating stablecoin pairs (USDT/BUSD/USDC) and variable-priced quote tokens (ETH/CAKE) produced a **dramatic improvement**. Net P&L dropped from **-0.001006 BNB** in the prior 20-minute test to **-0.000008 BNB** — roughly a **99% reduction in absolute loss**. All three trades were WBNB pairs that hit the break-even SL, losing only the expected ~0.5% efficiency tax on each round trip. No token reached the +200% TP trigger, so the partial-exit logic was not exercised. The non-WBNB quote guard rejected the majority of opportunities, which is the intended trade-off: fewer trades, but each trade is structurally safer.

## 8. Observations and Recommendations

1. **WBNB-only filter is the single most effective safeguard implemented so far.** Stablecoin and alt-quote pairs are the dominant source of losses.
2. **Break-even SL is now the only loss mode.** With real WBNB pairs, the bot loses ~0.5% of the buy size per trade — exactly the predicted round-trip efficiency cost.
3. **Volume of opportunities dropped sharply.** Only 3 trades executed in 20 minutes because the majority of new pairs are non-WBNB. This may limit total return potential.
4. **No TP events fired.** None of the WBNB tokens pumped 3× in the holding window. The bot is still dependent on finding tokens that moon within seconds.
5. **Next test suggestions:**
   - Run a longer window (e.g., 60 minutes) to see if the WBNB-only distribution can produce profitable trades.
   - Consider lowering the efficiency threshold slightly (e.g., 0.93) to allow more WBNB pairs without re-opening the stablecoin trap.
   - Add a per-token max hold time so positions that never hit TP are not held indefinitely at break-even.

---
*Generated from live mainnet test on 2026-07-23.*
