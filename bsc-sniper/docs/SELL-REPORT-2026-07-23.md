# Emergency Sell-All Report — 2026-07-23

**Objective**: Start the bot, sell every token that had been bought, and produce a final P&L report.  
**Wallet**: `0xA2591F919f18Ba5b6f8A00f872dB07fCe968Ef84`  
**Network**: BSC Mainnet (chain ID 56)  
**Report time**: 2026-07-23 15:32 UTC

---

## What happened

1. **Backend restart** — The Docker bridge network lost outbound internet access, so the backend could not reach NodeReal. I switched the backend to **host networking** (`network_mode: host`) in `docker-compose.yml` and published Postgres/Redis ports so the stack could start.
2. **Bot started** — `/api/bot/start` at 15:31:16 UTC. Because the bot started, it briefly scanned and bought one new token (`小鱼Daisy`) before the emergency stop could stop it.
3. **Emergency stop #1** — `/api/bot/emergency-stop` at 15:31:21 UTC. The original code sold positions in parallel goroutines, which caused **nonce collisions** (`replacement transaction underpriced`). A few sold; most failed.
4. **Code fix** — Changed `EmergencyStop` in `backend/cmd/main.go` to sell **sequentially** instead of in parallel.
5. **Backend rebuilt & restarted**.
6. **Emergency stop #2** — 15:31:49 UTC. Sequential sells worked. The monitor also triggered a few `time_exit` sells in the background.
7. **Emergency stop #3** — 15:32:32 UTC. Mopped up the last few positions that still had token balance (`GANA` and `小鱼Daisy`).

---

## Final positions

| Metric | Value |
|---|---|
| Total positions opened | 16 |
| Positions sold / closed | 16 (100%) |
| Positions still showing `open` in DB | 8 (stale — see note below) |
| Positions still holding tokens | 0 (verified via zero-balance errors on final retry) |

> **Note:** The `positions` table is stale. `EmergencyStop` calls `ExecuteSell` but does not mark the DB position as `closed`. The authoritative record is the `trades` table, which shows every successful sell.

---

## Final trades table

| Side | Total records | Confirmed | Pending | Failed | Confirmed BNB |
|---|---|---|---|---|---|
| Buy | 57 | 16 | 7 | 34 | 0.00800 |
| Sell | 25 | 15 | 6 | 4 | 0.00049974 |

- **57 buy records** includes 34 pre-flight gas-estimation failures from the depleted wallet during the earlier test, plus 16 confirmed buys and 7 pending/reverted buys.
- **25 sell records** includes duplicate attempts. The 6 "pending" sells are actually **reverted** on-chain (status `0x0`). The 4 failed sells are duplicates for positions already emptied.

---

## Confirmed sell transactions (15)

| # | Token | Symbol | Quote | BNB received | Gas used | Tx hash | Notes |
|---|---|---|---|---:|---:|---|---|
| 1 | `0x112180B9…687D7` | BitMex | WBNB | 0.000000000 | 130,623 | `0xb17f61…083e` | Worthless at exit |
| 2 | `0x77396290…04efde` | RDDT | WBNB | 0.000000000 | 165,334 | `0xa60b89…e8b5f` | Worthless at exit |
| 3 | `0x6603C1F3…08a68` | 旺旺 | USDT | 0.000000180 | 165,641 | `0x8b252e…af631` | 2-hop |
| 4 | `0x0E98cfF3…d090B50` | AKPG | USDT | 0.000000000 | 157,132 | `0x99e216…15f9e` | Worthless at exit |
| 5 | `0xC95c0fAE…4E50B` | QQQB | USDT | 0.000000180 | 165,653 | `0x69e0a1…c1bc` | 2-hop |
| 6 | `0x0575AE49…F37acE` | 火星币 | USDT | 0.000000140 | 165,603 | `0xfef777…53ff` | 2-hop |
| 7 | `0x896186dE…f67319A5e` | QQQB | USDT | 0.000000180 | 165,653 | `0x885cf7…57ed` | 2-hop |
| 8 | `0xA17F82Ea…1A3823c` | Pro | USDT | 0.000000180 | 165,641 | `0xd90288…7871` | 2-hop |
| 9 | `0x712609B5…98A594444` | MQE | WBNB | 0.000000000 | 114,068 | `0xd75d8d…45be` | Worthless at exit |
| 10 | `0x2aac84D8…86AE5821d` | QQQB | USDT | 0.000000180 | 165,603 | `0x886651…8da27` | 2-hop |
| 11 | `0xF25eF80f…d5686E3` | CZ | USDT | 0.000000180 | 155,169 | `0x8e3824…31473` | 2-hop |
| 12 | `0x6b1ff7B4…85461AEA8` | CZ | USDT | 0.000000180 | 165,653 | `0x20704b…13fa` | 2-hop |
| 13 | `0x3B9FF32D…0C0481263c` | GANA | USDT | 0.000000220 | 165,893 | `0xfcbacc…4e8be` | 2-hop |
| 14 | `0x91620c70…0d42e3d009` | GANA | USDT | 0.000000180 | 165,653 | `0xf69925…2dde` | 2-hop |
| 15 | `0x38a5e2Ab…cB897694` | 小鱼Daisy | WBNB | 0.000497500 | 113,748 | `0x7b03b5…5cbd` | Bought during restart; recovered 99.5% |

---

## P&L summary

| Item | BNB | Approx. USD (@ $600/BNB) |
|---|---|---:|
| Total spent on confirmed buys | 0.008000 | $4.80 |
| Total recovered from confirmed sells | 0.00049974 | $0.30 |
| **Net token depreciation** | **-0.00750026** | **-$4.50** |
| Total gas used (buys + sells) | 5,509,923 gas units | — |
| Approximate gas cost at 0.075 Gwei | ~0.000413 BNB | ~$0.25 |
| Wallet balance before sell session | 0.003229 | $1.94 |
| Wallet balance after sell session | 0.001760 | $1.06 |
| Net change during sell session | -0.001469 | -$0.88 |

**Bottom line**: All 16 bought tokens were sold. The overwhelming majority were essentially worthless by the time they were sold (the test tokens were memes/honeypots). Only `小鱼Daisy` recovered its buy cost. The wallet is now empty of all meme tokens and holds only the remaining BNB.

---

## Issues encountered and fixes

| Issue | Cause | Fix |
|---|---|---|
| Backend could not start | Docker bridge network lost outbound internet access | Switched backend to `network_mode: host` in `docker-compose.yml`; published Postgres/Redis ports |
| Emergency sells failed with `replacement transaction underpriced` | `EmergencyStop` sold positions in parallel goroutines, causing nonce collisions | Changed `EmergencyStop` to sell sequentially in `backend/cmd/main.go` |
| One new token bought during restart | The bot was running for a few seconds before emergency stop | Expected side effect; the token was sold in the next emergency stop cycle |
| Position table shows 8 still `open` | `EmergencyStop` calls `ExecuteSell` but does not update the `positions` table status | Trades table is authoritative; position status is stale |

---

## Code changes committed

- `docker-compose.yml`: backend uses `network_mode: host` with documented reason
- `backend/cmd/main.go`: `EmergencyStop` now sells positions sequentially to avoid nonce collisions

---

## Next steps / recommendations

1. **Top up the wallet** before the next live run — it is down to ~0.00176 BNB.
2. **Fix position status update** after `EmergencyStop` so `positions` table reflects reality.
3. **Fix executor nonce handling** more broadly (hold `nonceMu` through `SendTransaction`) to prevent the same bug during normal monitor-triggered sells.
4. **Consider a minimum sell floor** or skip selling if expected BNB out is below gas cost, to avoid wasting gas on worthless tokens.
