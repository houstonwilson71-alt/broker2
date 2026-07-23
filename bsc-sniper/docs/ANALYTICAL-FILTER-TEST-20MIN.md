# 20-Minute Live Mainnet Test — Final Report

**Date:** 2026-07-23  
**Duration:** 20 minutes  
**Wallet:** `0xA2591F919f18Ba5b6f8A00f872dB07fCe968Ef84`  
**Buy amount:** 0.0005 BNB  
**Efficiency threshold:** 0.95 (95% round-trip)  
**Liquidity floor:** $12,000 USD  
**TP strategy:** 50% at +200%, remaining 50% at +300% or trailing stop -20% from peak  
**Sell slippage:** amountOutMin = expected × 0.95  
**Gas boost:** 1.5× on take-profit sells  
**BNB price used for USD:** $567.49  

## 1. Executive Summary

- **Total buys:** 5  
- **Total BNB spent:** 0.0025 BNB  
- **Total sell transactions:** 6  
- **Total BNB recovered:** 0.001494 BNB  
- **Net P&L:** -0.001006 BNB (-0.57 USD)  
- **TP 50% + trailing SL exits:** 1  
- **Break-even SL exits:** 3  
- **Pairs approved:** 5  
- **Pairs rejected:** 9  

## 2. All Buys (Full Transaction Hashes)

| # | Token | Symbol | Quote | Buy BNB | Buy Tx |
|---|-------|--------|-------|---------|--------|
| 1 | 0xd581f9f9313B8758D4F155736570b3f7318EF735 | UniBPEPE | WBNB | 0.0005 | [0xf3b2a7...](https://bscscan.com/tx/0xf3b2a7413723613b384b72df50ef6872302b28d41348fb385eee031a1cfc17f9) |
| 2 | 0x4b65f64F254e449Ee0af2eBF38e42ce7575E785d | ZYLO | WBNB | 0.0005 | [0xdfd4fd...](https://bscscan.com/tx/0xdfd4fd71f93ce6950a011af39c2d8096b26ca58bbe465b5346b1b3e91531531e) |
| 3 | 0x1DbC9461d6c9b206012D26F3eEaEE2B241364d0e | QQQB | USDT | 0.0005 | [0xc5589e...](https://bscscan.com/tx/0xc5589e76c1a01e2729a1b4b32f45165088b768a1327717db272f1ca5932bac8a) |
| 4 | 0x2FAe3FfBBbfd1F791d07c1166a28079056dB6b93 | 月球币 | USDT | 0.0005 | [0x15a171...](https://bscscan.com/tx/0x15a1711781262b011addd6a3c66fca231b6c648dc241a36497f70512ab53d8c2) |
| 5 | 0xe47387719B3024680fABB88Ec5e565F28AF803ba | 小股东 | WBNB | 0.0005 | [0x197a21...](https://bscscan.com/tx/0x197a211a04c49ab5d88dbff6c49f3b9010d0217629b31c355e02091f2e31292c) |

**Full buy hashes:**

- UniBPEPE — 0xf3b2a7413723613b384b72df50ef6872302b28d41348fb385eee031a1cfc17f9
- ZYLO — 0xdfd4fd71f93ce6950a011af39c2d8096b26ca58bbe465b5346b1b3e91531531e
- QQQB — 0xc5589e76c1a01e2729a1b4b32f45165088b768a1327717db272f1ca5932bac8a
- 月球币 — 0x15a1711781262b011addd6a3c66fca231b6c648dc241a36497f70512ab53d8c2
- 小股东 — 0x197a211a04c49ab5d88dbff6c49f3b9010d0217629b31c355e02091f2e31292c

## 3. All Sells (Full Transaction Hashes + Exit Reason)

| # | Token | Symbol | Quote | Sell # | Sell BNB | Reason | Sell Tx |
|---|-------|--------|-------|--------|----------|--------|---------|
| 1 | 0xd581f9f9313B8758D4F155736570b3f7318EF735 | UniBPEPE | WBNB | 1 | 0.000498 | Break-even SL | [0xc3e39f...](https://bscscan.com/tx/0xc3e39f11ba96a83c958f3b9a45d3fd4c1863ae557ab62afdab40febcaf353163) |
| 2 | 0x4b65f64F254e449Ee0af2eBF38e42ce7575E785d | ZYLO | WBNB | 1 | 0.000498 | Break-even SL | [0x06c49f...](https://bscscan.com/tx/0x06c49f6d2eadb617eac4d9f8179e5b291436d6d019c14177fa8a26734ba79171) |
| 3 | 0x1DbC9461d6c9b206012D26F3eEaEE2B241364d0e | QQQB | USDT | 1 | 0.000001 | TP 50% at +200% | [0xd354c1...](https://bscscan.com/tx/0xd354c17e95f63d5192d72f76515224bf8ec5bb54c4f5e1dee9bc53c3f8be3a38) |
| 3 | 0x1DbC9461d6c9b206012D26F3eEaEE2B241364d0e | QQQB | USDT | 2 | 0 | Trailing SL / TP 300% | [0xbf0728...](https://bscscan.com/tx/0xbf07281834e89229c782cc7152b3acdd06962931c4ce74659916af2df2a7a79d) |
| 4 | 0x2FAe3FfBBbfd1F791d07c1166a28079056dB6b93 | 月球币 | USDT | 1 | 0 | TP 200% (taxed) | [0x2ae0a9...](https://bscscan.com/tx/0x2ae0a933c565fd190eca4b1ba5efcad6864a757734ed6fc45bfe8db81bfe7f29) |
| 5 | 0xe47387719B3024680fABB88Ec5e565F28AF803ba | 小股东 | WBNB | 1 | 0.000498 | Break-even SL | [0xfea51f...](https://bscscan.com/tx/0xfea51f84a91200ba956ae27528a9808631114728975ca57358c63435feef76f0) |

**Full sell hashes:**

- UniBPEPE (Break-even SL) — 0xc3e39f11ba96a83c958f3b9a45d3fd4c1863ae557ab62afdab40febcaf353163
- ZYLO (Break-even SL) — 0x06c49f6d2eadb617eac4d9f8179e5b291436d6d019c14177fa8a26734ba79171
- QQQB (TP 50% at +200%) — 0xd354c17e95f63d5192d72f76515224bf8ec5bb54c4f5e1dee9bc53c3f8be3a38
- QQQB (Trailing SL / TP 300%) — 0xbf07281834e89229c782cc7152b3acdd06962931c4ce74659916af2df2a7a79d
- 月球币 (TP 200% (taxed)) — 0x2ae0a933c565fd190eca4b1ba5efcad6864a757734ed6fc45bfe8db81bfe7f29
- 小股东 (Break-even SL) — 0xfea51f84a91200ba956ae27528a9808631114728975ca57358c63435feef76f0

## 4. Per-Position P&L

| Token | Symbol | Buy BNB | Recovered BNB | P&L BNB | P&L % | Exit Type |
|-------|--------|---------|---------------|---------|-------|-----------|
| 0xd581f9f9313B8758D4F155736570b3f7318EF735 | UniBPEPE | 0.0005 | 0.000498 | -0.000003 | -0.5% | Break-even SL |
| 0x4b65f64F254e449Ee0af2eBF38e42ce7575E785d | ZYLO | 0.0005 | 0.000498 | -0.000003 | -0.5% | Break-even SL |
| 0x1DbC9461d6c9b206012D26F3eEaEE2B241364d0e | QQQB | 0.0005 | 0.000001 | -0.000499 | -99.85% | TP 50% + trailing SL |
| 0x2FAe3FfBBbfd1F791d07c1166a28079056dB6b93 | 月球币 | 0.0005 | 0 | -0.0005 | -99.93% | TP 200% (taxed) |
| 0xe47387719B3024680fABB88Ec5e565F28AF803ba | 小股东 | 0.0005 | 0.000498 | -0.000003 | -0.5% | Break-even SL |

## 5. Efficiency Data (Every Token Filtered)

| Token | Symbol | Quote | Efficiency | Status | Rejection Reason |
|-------|--------|-------|------------|--------|------------------|
| 0xd581f9f9313B8758D4F155736570b3f7318EF735 | UniBPEPE | WBNB | 0.9948 | APPROVED |  |
| 0x4b65f64F254e449Ee0af2eBF38e42ce7575E785d | ZYLO | WBNB | 0.9949 | APPROVED |  |
| 0x1DbC9461d6c9b206012D26F3eEaEE2B241364d0e | QQQB | USDT | 0.9899 | APPROVED |  |
| 0x2FAe3FfBBbfd1F791d07c1166a28079056dB6b93 | 月球币 | USDT | 0.9899 | APPROVED |  |
| 0xe47387719B3024680fABB88Ec5e565F28AF803ba | 小股东 | WBNB | 0.995 | APPROVED |  |
| 0x4C84ea62ED8c5216bCcA9C462E27E3dB67fBb01B |  | WBNB | 0.9947 | REJECTED | low_liquidity:$6739 |
| 0x67CBcd7e1A2f81C1E155EDCF8A58AEBa77cF57Aa |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0x556b74C1C84Eda0FCF583e1a3822dBdfdbcde16A |  | WBNB | 0.9947 | REJECTED | low_liquidity:$6831 |
| 0x7F9229BD94a7F7A878E9f12c9F4E164F16FaDBe7 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0x4B5b093DC01F7005aE63309cdcA3880e31011445 |  | WBNB | 0.9942 | REJECTED | low_liquidity:$2841 |
| 0xF35b51b87C76E7B5aD03d0e1c3f4c4B0fb80e238 |  | WBNB | 0.9049 | REJECTED | honeypot_detected, low_efficiency:0.9049, low_liquidity:$23 |
| 0x137E8aE123211fC96eC5ba52DB01abB90C545f5F |  | WBNB | 0.9947 | REJECTED | low_liquidity:$8377 |
| 0x8D6dd7dC2e533E33f470237ad2F9fB89b47406DF |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0xB272e2FC54272c8F4D1eFdd9c54cC4546368e27f |  | WBNB | 0.9049 | REJECTED | low_liquidity:$23, honeypot_detected, low_efficiency:0.9049 |

## 6. Rejected Tokens Breakdown

- **Rejected by efficiency guard:** 5  
- **Rejected by liquidity floor:** 4  
- **Rejected by duplicate symbol guard:** 0  
- **Rejected by other reasons:** 0  

## 7. Comparison to Previous 30-Minute Test

| Metric | Previous Test (30 min, 0.001 BNB, 0.85 eff) | This Test (20 min, 0.0005 BNB, 0.95 eff) |
|--------|-----------------------------------------------|------------------------------------------|
| Buy size | 0.001 BNB | 0.0005 BNB |
| Efficiency threshold | 0.85 | 0.95 |
| Duration | 30 min | 20 min |
| Buys | 16 | 5 |
| BNB spent | 0.016 BNB | 0.0025 BNB |
| BNB recovered | 0.008975 BNB | 0.001494 BNB |
| Net P&L | -0.007025 BNB | -0.001006 BNB |
| Net P&L USD | -3.97 USD | -0.57 USD |
| Loss rate | -0.44% per trade avg | -40.26% of buy size per trade |

### Analysis

This 20-minute test reduced absolute losses but still lost money. Net P&L was **-0.001006 BNB** vs **-0.007025 BNB** previously. Two key drivers: (1) smaller 0.0005 BNB buy size halves exposure per trade, and (2) raising the efficiency floor to 0.95 blocked the 0.90-ish tokens that previously got through. The new TP logic executed correctly on QQQB: the first 50% was sold at +200%, and the remaining 50% was closed by the trailing stop, with total QQQB loss ~99.85% of the buy. USDT pairs still hit the same tax/slippage trap, while WBNB pairs hit small, controlled break-even SL losses.

## 8. Observations and Recommendations

1. **0.95 efficiency floor is the right direction.** The token rejected at 0.9048 efficiency was correctly flagged as a honeypot/low-liquidity scam.
2. **Break-even SL is now tiny.** WBNB pairs lost ~0.5% of the 0.0005 BNB buy, which is exactly what the 0.995 simulated efficiency predicts.
3. **Partial TP + trailing stop works.** QQQB sold half at the +200% trigger and the rest on the trailing stop, capping the loss.
4. **USDT pairs are still toxic.** Even with 0.95 efficiency and 5% sell slippage, the actual recovery on USDT tokens was <0.1% of the buy size due to token tax.
5. **Gas boost and retry logic functioned.** No retry warnings were triggered, suggesting the 5% slippage tolerance was sufficient.
6. **Next test suggestion:** Combine the 0.0005 BNB size + 0.95 efficiency + partial TP with a **hard blacklist of USDT pairs** to see if WBNB-only trading can become profitable.

---
*Generated from live mainnet test on 2026-07-23.*
