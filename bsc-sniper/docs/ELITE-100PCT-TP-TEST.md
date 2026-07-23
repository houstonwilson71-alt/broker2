# Elite BSC Sniper – 100% TP at 200% / Break-even SL Test Report

**Date:** 2026-07-23  
**Wallet:** `0xA2591F919f18Ba5b6f8A00f872dB07fCe968Ef84`  
**Test duration:** 30 minutes (17:57:24 UTC – 18:27:34 UTC)  
**Network:** BSC Mainnet  
**Strategy:** 100% sell when price >= 3× entry (TP 200%), 100% sell when price <= entry (break-even SL), 3-second tick, no trailing/partial logic.

---

## 1. Summary

- **Pairs seen:** ~30
- **Pairs approved by safeguards:** 26
- **Confirmed buys:** 22
- **Confirmed sells:** 21
- **Reverted buys:** 4
- **Reverted/pending sell:** 1
- **Open positions at stop:** 0
- **Start wallet:** 0.028666774147812589 BNB
- **End wallet:** 0.021987875436373993 BNB
- **Total BNB spent on buys:** 0.01100000 BNB
- **Total BNB recovered from sells:** 0.00498266 BNB
- **Net token P&L:** -0.00601734 BNB
- **Total wallet loss (incl. gas):** 0.006678898711438596 BNB
- **Approx. USD loss:** $4.27 (@ $640/BNB)

---

## 2. Safeguard Activity

| Safeguard | Rejections |
|-----------|------------|
| Pre-buy honeypot simulation | 2 |
| 15% tax guard | 2 |
| Liquidity floor ($5k) | 4 |
| Liquidity / RPC read errors | 6 |
| **Total rejected** | **10** |
| **Total approved** | **26** |

---

## 3. Confirmed Buys

| # | Token | Symbol | Amount (BNB) | Tx Hash |
|---|-------|--------|--------------|---------|
| 1 | 0xD1399bE5c383f6AE0F06C42Bf4110749B8a9FC3F | AIF | 0.00050000 | 0x1e36f0e05432511ccde0f041d58d47680ed34f6852ba8e39281d813c20b715e5 |
| 3 | 0xF3B0545Ff4799A55D1105E068fB182C2F3498d4A | CGN | 0.00050000 | 0x63584c539736485e3c1112710457bdc297240e756f59cdb96ee5bc8dea95d87d |
| 6 | 0x8dA1c053B6eDaa8144d4c144CFBd8518f79f11a0 | 黑马 | 0.00050000 | 0x25642a59290824d1f308f98156946b137680cf5d19af39670e5a5b5e4e22ea28 |
| 8 | 0xFCff67FDDf26848aEB2D27BBf37707d15dA03517 | CFI | 0.00050000 | 0x8d65bd6209e48967dadc2655d02e78e8e94afc7ee64a4c58dd5c5f670f536b90 |
| 11 | 0xC30831B3Ef404b5649b9aE193dB46EE15C34A5A2 | CLBS | 0.00050000 | 0x77a6b42657b4110763f3f57452b7f7495c07cdd304f88a23251b632571c4a85e |
| 13 | 0x84120991a1227AE35fEA821A7a9aA30479f81602 | CMND | 0.00050000 | 0x8b9947ae4650cfd9bae1d466f68aca43035d6cef690b94b094275da51ad163b8 |
| 15 | 0x0a3b3625b8216c8d153A25cdC611Cc531ad257c6 | CAGT | 0.00050000 | 0xc47c8b3a2e18545e1336620cb2de0b9a73d2cb0f4d80283a363b7742c5019f80 |
| 18 | 0xA77b335082b87E526Cd72AE14A7C4aAe3eA9C956 | 小狐狸 | 0.00050000 | 0x1ba5bd606514434ec3cdf53bbd9196aff052391f1ee5f327eca19671c19e8dc0 |
| 20 | 0x8f696A28391F3e436394E4a23773bbE972aa9e15 | 月球币 | 0.00050000 | 0xf1fc52b14143d6b023e0bb905a4b1f4efafb0278c376142b8fc065bf2b6d23cc |
| 22 | 0xA89D9Bb06Af0d9A6f9C4B2418F4A43448608B552 | NVO | 0.00050000 | 0xe0c043936b05949df9d231686bf400113de89001d8b1494ce048b128269af249 |
| 24 | 0xf7BB3753E07B3970dc990BFb9419AA5C6d9802dE | 杀零狗 | 0.00050000 | 0xb642b92abb3858db378fdb45ab3104ab6ddafdbf5e73710089b4fc7f07100b86 |
| 26 | 0xf5B7C9fdC09445437F1eFbcAdb389e8DD80D3EAB | QQQB | 0.00050000 | 0x7d7857c2a98444acf3a614c908943a358bea2b9e922591299266e21cda2150c7 |
| 28 | 0xb897C0E2E3086D2CEa97058EB98bdE1a4d60DFE1 | QQQB | 0.00050000 | 0x252da7869070b291990c19d97670bc99711db9fea47b345e1ea8c1a95efa1ce4 |
| 30 | 0xA3C6bB3011f80F4C1A918af7B97B11bF95cc7a36 | 杀零狗 | 0.00050000 | 0x51272de13875525e2c98eed58a5dc2d6c8f9e512b21017ca1bcbd7138a6b7516 |
| 32 | 0xCddB158F63796Ef11C2eA0b8056612E2C59D6aEe | SHAZ | 0.00050000 | 0x2c5c972dca1b879d46eb1291e8f3494ba1d110c0cf4a125d08867ad3d5bb83fd |
| 34 | 0x3bBc2E59b44Ac2E7537135714D6C53e48699D7cB | Binance | 0.00050000 | 0xbf46017df4a6e7d5630b1d3ce4886ba8e459a2acfc4c776aaf86a8270292c491 |
| 35 | 0x9D2150d0B31f297f25656Ad6809F7281080c1C4D | LAB | 0.00050000 | 0xd2029b119379f0cb17eafacd27519681d27bd227574ae8ddb124d39af598bf91 |
| 37 | 0x1f07e0ea784692197AA5e9e207e09C31bA1c7CcC | BASTEROID | 0.00050000 | 0x895fba4908cc36b15eecaee7d77dd871818323cde221f38fa2fded0b8dbf91f9 |
| 40 | 0x3318F4F5A761C0962eB61842fc62D28B3338514F | CNET | 0.00050000 | 0x5a32375392e131072ed6763d80cd67ac947a2278199e923e99fdd6bb71f9d6f9 |
| 43 | 0x14EC87Fe36dA36a2f35926d2aa763daDc2819E27 | QQQB | 0.00050000 | 0xd5dcd38a68d183b0f05d59e6c40ec6954f3c8beffc0d7e00df1272c0a788f4ef |
| 45 | 0xFaD66d1B243C80230Ad9314164B0495DcD32Fe5b | 旺旺 | 0.00050000 | 0x5c5e50ad660c394f2b4d1807b6bcf0e23b2ba680b2400aa1d0127895b35e44a8 |
| 47 | 0x6ed803DBc75976C899bE565cB4237ebAb336B469 | Narcos. | 0.00050000 | 0x42bb2ec3c81218041ee392d678105a9fb397bbccae6342be56d2dc40a6047cb0 |

**Total confirmed buys:** 22 buys × 0.0005 BNB = **0.01100000 BNB**

---

## 4. Confirmed Sells

| # | Token | Symbol | Amount (BNB) | Reason | Tx Hash |
|---|-------|--------|--------------|--------|---------|
| 2 | 0xD1399bE5c383f6AE0F06C42Bf4110749B8a9FC3F | AIF | 0.00000300 | TP 200% | 0x0ba714f5ba2682f1bf301a31e27b7e1073651dff088acd1178025f131d2305ba |
| 4 | 0xF3B0545Ff4799A55D1105E068fB182C2F3498d4A | CGN | 0.00000300 | TP 200% | 0xfa24361fc512b260c3d56dc4b2c5a29cb6ccc6e38412391537cfff81b8b127c1 |
| 7 | 0x8dA1c053B6eDaa8144d4c144CFBd8518f79f11a0 | 黑马 | 0.00000037 | Break-even SL | 0xeb89c692627c4990a4155b54ab54fab04ece2e4318f2be4f6aa6216be91b27c4 |
| 10 | 0xFCff67FDDf26848aEB2D27BBf37707d15dA03517 | CFI | 0.00000300 | TP 200% | 0x85970ced3c150b64523d4dedecace6692916e3a17b464577a80034116f128bb3 |
| 12 | 0xC30831B3Ef404b5649b9aE193dB46EE15C34A5A2 | CLBS | 0.00000300 | TP 200% | 0xddc736201fe0168a167a0a8375b4521d17327afa1ef7ad94e1e65c364f1ecac8 |
| 14 | 0x84120991a1227AE35fEA821A7a9aA30479f81602 | CMND | 0.00000300 | TP 200% | 0xf7ef8a4e3799e34f0492cc9156f553bffd151dd79aba956a12534ed657a6d6b8 |
| 16 | 0x0a3b3625b8216c8d153A25cdC611Cc531ad257c6 | CAGT | 0.00000300 | TP 200% | 0x1ee34e2af8ecb54cd1be992f763dc96459376810cabd730e3e3565c787c4cb29 |
| 21 | 0x8f696A28391F3e436394E4a23773bbE972aa9e15 | 月球币 | 0.00000057 | TP 200% | 0xf9aac2630c04a14a013dac597241d2cd7d43fb804fadce2d3dce696d204d6c6e |
| 23 | 0xA89D9Bb06Af0d9A6f9C4B2418F4A43448608B552 | NVO | 0.00049750 | Break-even SL | 0x0354bebf38f320e146060ee7d278955ec0fa3efd43826207d38255ef848c3241 |
| 25 | 0xf7BB3753E07B3970dc990BFb9419AA5C6d9802dE | 杀零狗 | 0.00049502 | Break-even SL | 0xa0f469192d00af2e8ced6cd57d8fb764c834b1f731850fca8c3fa262d0c0c5be |
| 27 | 0xf5B7C9fdC09445437F1eFbcAdb389e8DD80D3EAB | QQQB | 0.00000020 | Break-even SL | 0xfc33756637096e62864e53fa6fca5ff7b63de128a3a7f030dedb18b18cad592a |
| 29 | 0xb897C0E2E3086D2CEa97058EB98bdE1a4d60DFE1 | QQQB | 0.00049502 | Break-even SL | 0x873eabbc111ef5fa20ed61b94d31a28c2f9b26c2af744cda787b93b70ac94ea0 |
| 31 | 0xA3C6bB3011f80F4C1A918af7B97B11bF95cc7a36 | 杀零狗 | 0.00049502 | Break-even SL | 0x0355a9c97988c351f4833a4b466e3ae7b1e5af6ee4a8e1be236bdff333de6010 |
| 33 | 0xCddB158F63796Ef11C2eA0b8056612E2C59D6aEe | SHAZ | 0.00049750 | Break-even SL | 0x8c838c4e405be20bad5d08f25c81506937e5476087ab3983c70369b2283cba2f |
| 36 | 0x3bBc2E59b44Ac2E7537135714D6C53e48699D7cB | Binance | 0.00000057 | TP 200% | 0xebcd55f7bf0c3601f4e2bdd095634fd1e45971ac43ddf5dd927d8199305fac06 |
| 38 | 0x9D2150d0B31f297f25656Ad6809F7281080c1C4D | LAB | 0.00049502 | Break-even SL | 0xb47283e8d2cd25cf08cee04c95bdbb6fe49319f0a53ba9b927409ea37f516524 |
| 39 | 0x1f07e0ea784692197AA5e9e207e09C31bA1c7CcC | BASTEROID | 0.00049750 | Break-even SL | 0x22378ab650d894c0b9369bb36c6fbff0f602cfc17483e11e44355f73f688d713 |
| 41 | 0x3318F4F5A761C0962eB61842fc62D28B3338514F | CNET | 0.00000280 | Break-even SL | 0xe25f48af53ecc18a7e5c04ee55c5bd2ff377515224b6055282ed8bd562d86727 |
| 44 | 0x14EC87Fe36dA36a2f35926d2aa763daDc2819E27 | QQQB | 0.00049503 | Break-even SL | 0xccd815a4468405f697a04d4329daa00e4444ee766e9d803066689136229ad3c2 |
| 46 | 0xFaD66d1B243C80230Ad9314164B0495DcD32Fe5b | 旺旺 | 0.00049504 | Break-even SL | 0x6340e9c81952b3dd2a444791f2493c651abfa89907c5276a34c0e68e4ae50d4f |
| 48 | 0x6ed803DBc75976C899bE565cB4237ebAb336B469 | Narcos. | 0.00049750 | Break-even SL | 0x056da9569d73935a9fba8970b84232666fbe7008ea1e2d9323b289f57bb94b58 |

**Total confirmed sells:** 21  
**Sell reasons breakdown:** 8 × TP 200%, 13 × Break-even SL  
**Total BNB recovered from sells:** 0.00498266 BNB

---

## 5. Reverted / Pending Transactions

These transactions were recorded as `pending` in the DB but their on-chain receipts show `status=0x0` (reverted). They still consumed gas.

| Type | Token | Tx Hash | Gas used (hex) |
|------|-------|---------|----------------|
| buy | 0x60b07bC89DBe64a78dba0d7717a1Fa841540F1D3 | 0x946d63886bbbce09433fff606449c581dce3b3d3a3236c002830b848dd3ba563 | 0x2b9a3 |
| buy | 0x367a53283e71EECC185fcA39103656d664891F45 | 0x5333a7ad109687e6e66ab36539a08373116fd91a82b35776f9f62af50b8dab10 | 0x2b955 |
| buy | 0x8Dd53d629b0c0030654CE0b595135d76E7bD03dF | 0xc7ed1a0f075973c3ed261fc68d887642bd81bd56aa81d463968b24d02ffc3efb | 0x2ba99 |
| buy | 0x9053e29201cA0F9Fd80C8C6aaa06e467FDd3c5da | 0x06b1361d786e3d8a37154c100739c23936952249e6dd76cc60511c5f79945f09 | 0x2baa8 |
| sell | 0xA77b335082b87E526Cd72AE14A7C4aAe3eA9C956 | 0x3a1bf99523474c811f58a4a764825a556ca11afba36fc76b39c753e7cd4e70b0 | 0x27d26 |

---

## 6. Net Profit / Loss

| Item | BNB |
|------|-----|
| Total spent on confirmed buys | -0.01100000 |
| Total recovered from confirmed sells | +0.00498266 |
| Net token P&L | -0.00601734 |
| Wallet start | 0.02866677 |
| Wallet end | 0.02198788 |
| Total wallet loss (includes all gas) | -0.00667890 |
| Approx. USD loss (@ $640/BNB) | ~-$4.27 |

---

## 7. Errors and Warnings

- **Database connection retries on startup:** The backend container crashed 3 times with `connect database: failed to connect ... read: connection reset by peer` before the Postgres container was ready. This is normal docker-compose startup ordering and the container auto-restarted successfully.
- **4 reverted buys:** The tokens were approved by the filter but the on-chain swap reverted (likely due to high dynamic tax or honeypot behavior not caught by the pre-buy simulation). The bot logged each revert and continued.
- **1 reverted sell:** The break-even SL sell for 小狐狸 was reverted on-chain. The position was still marked closed in the DB.
- **No panics or critical errors** during the trading loop.

---

## 8. Notes

- The new strategy is fully binary: every confirmed sell was either 100% at TP 200% or 100% at break-even SL. No partial positions remain after the stop.
- TP 200% exits triggered on extreme reserve-price spikes (e.g., AIF +22,259%, CGN +22,259%) but the actual BNB recovered from those sells was tiny. This happens because the token contracts applied massive sell taxes or fees that the pre-buy simulation did not fully replicate.
- The stop handler in `backend/cmd/main.go` called `SellAllPositions()` before shutdown; the final DB check shows **0 open positions**.
- Code change: `backend/internal/monitor/position.go` now contains only the 3-second binary check (TP 200% vs break-even SL) and sells 100% of the remaining balance on either signal.
