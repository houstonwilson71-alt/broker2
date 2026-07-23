# Comparative Analysis: 22 Confirmed Trades (30-Minute Mainnet Test)

**Date:** 2026-07-23  
**Wallet:** `0xA2591F919f18Ba5b6f8A00f872dB07fCe968Ef84`  
**Strategy:** 100% TP at 200% (price ≥ 3× entry) OR 100% break-even SL (price ≤ entry)  
**Data source:** Local PostgreSQL `trades`, `positions`, `filter_results`, and `tokens` tables.

> **Note:** The live log files (`logs/bot.log`) were not present in the workspace at analysis time, so this report is built entirely from the database. On-chain details such as exact block numbers, holder counts, and true transfer-tax percentages were not queried externally because no BSCScan API key or archive RPC was configured in the environment.

## 1. Executive Summary

- **Confirmed round-trip trades:** 22 (8 TP 200%, 13 break-even SL, 1 reverted sell)
- **Reverted buy attempts:** 4
- **Total tokens approved by filter:** 26
- **Total BNB spent on confirmed buys:** 0.01100000 BNB
- **Total BNB recovered from confirmed sells:** 0.00498266 BNB
- **Net token P&L:** -0.00601734 BNB
- **Average round-trip efficiency:** 0.4530 (sell BNB / buy BNB)

## 2. Raw Data Table — All Confirmed Round-Trips

| # | Token | Symbol | Quote | Liquidity (USD) | Entry Price (BNB) | Buy BNB | Sell BNB | PnL (BNB) | PnL % | Exit Type | Sell Price (BNB) | Time to Sell (s) | Buy Tx | Sell Tx |
|---|-------|--------|-------|-----------------|-------------------|---------|----------|-----------|-------|-----------|------------------|------------------|--------|---------|
| 1 | 0xD1399bE5c383f6AE0F06C42Bf4110749B8a9FC3F | AIF | WBNB | $8188.01 | 0.000007258645363409 | 0.0005 | 0.000003 | -0.000497 | -99.4% | TP 200% | 0.00000004355070185 | 14.2 | [0x1e36f0...](https://bscscan.com/tx/0x1e36f0e05432511ccde0f041d58d47680ed34f6852ba8e39281d813c20b715e5) | [0x0ba714...](https://bscscan.com/tx/0x0ba714f5ba2682f1bf301a31e27b7e1073651dff088acd1178025f131d2305ba) |
| 3 | 0xF3B0545Ff4799A55D1105E068fB182C2F3498d4A | CGN | WBNB | $8776.09 | 0.000000007779948622 | 0.0005 | 0.000003 | -0.000497 | -99.4% | TP 200% | 0.000000000046678436 | 14.6 | [0x63584c...](https://bscscan.com/tx/0x63584c539736485e3c1112710457bdc297240e756f59cdb96ee5bc8dea95d87d) | [0xfa2436...](https://bscscan.com/tx/0xfa24361fc512b260c3d56dc4b2c5a29cb6ccc6e38412391537cfff81b8b127c1) |
| 6 | 0x8dA1c053B6eDaa8144d4c144CFBd8518f79f11a0 | 黑马 | USDT | $14000 | 0.000000012412561517 | 0.0005 | 0.00000037 | -0.00049963 | -99.93% | Break-even SL | 0.000000000009307178 | 12.9 | [0x25642a...](https://bscscan.com/tx/0x25642a59290824d1f308f98156946b137680cf5d19af39670e5a5b5e4e22ea28) | [0xeb89c6...](https://bscscan.com/tx/0xeb89c692627c4990a4155b54ab54fab04ece2e4318f2be4f6aa6216be91b27c4) |
| 8 | 0xFCff67FDDf26848aEB2D27BBf37707d15dA03517 | CFI | WBNB | $6755.17 | 0.000000005985462406 | 0.0005 | 0.000003 | -0.000497 | -99.4% | TP 200% | 0.000000000035911809 | 14.3 | [0x8d65bd...](https://bscscan.com/tx/0x8d65bd6209e48967dadc2655d02e78e8e94afc7ee64a4c58dd5c5f670f536b90) | [0x85970c...](https://bscscan.com/tx/0x85970ced3c150b64523d4dedecace6692916e3a17b464577a80034116f128bb3) |
| 11 | 0xC30831B3Ef404b5649b9aE193dB46EE15C34A5A2 | CLBS | WBNB | $6800.44 | 0.000000000301278133 | 0.0005 | 0.000003 | -0.000497 | -99.4% | TP 200% | 0.00000000000180762 | 15.4 | [0x77a6b4...](https://bscscan.com/tx/0x77a6b42657b4110763f3f57452b7f7495c07cdd304f88a23251b632571c4a85e) | [0xddc736...](https://bscscan.com/tx/0xddc736201fe0168a167a0a8375b4521d17327afa1ef7ad94e1e65c364f1ecac8) |
| 13 | 0x84120991a1227AE35fEA821A7a9aA30479f81602 | CMND | WBNB | $6708.96 | 0.000000000297268108 | 0.0005 | 0.000003 | -0.000497 | -99.4% | TP 200% | 0.000000000001783561 | 13.9 | [0x8b9947...](https://bscscan.com/tx/0x8b9947ae4650cfd9bae1d466f68aca43035d6cef690b94b094275da51ad163b8) | [0xf7ef8a...](https://bscscan.com/tx/0xf7ef8a4e3799e34f0492cc9156f553bffd151dd79aba956a12534ed657a6d6b8) |
| 15 | 0x0a3b3625b8216c8d153A25cdC611Cc531ad257c6 | CAGT | WBNB | $6731.59 | 0.00000059654122807 | 0.0005 | 0.000003 | -0.000497 | -99.4% | TP 200% | 0.000000003579151194 | 13.3 | [0xc47c8b...](https://bscscan.com/tx/0xc47c8b3a2e18545e1336620cb2de0b9a73d2cb0f4d80283a363b7742c5019f80) | [0x1ee34e...](https://bscscan.com/tx/0x1ee34e2af8ecb54cd1be992f763dc96459376810cabd730e3e3565c787c4cb29) |
| 18 | 0xA77b335082b87E526Cd72AE14A7C4aAe3eA9C956 | 小狐狸 | USDT | $12000 | 0.00000001064016825 | 0.0005 | 0.00049502 | -0.0005 | -100% | Reverted | 0.000000010534166449 | 10.8 | [0x1ba5bd...](https://bscscan.com/tx/0x1ba5bd606514434ec3cdf53bbd9196aff052391f1ee5f327eca19671c19e8dc0) | [0x3a1bf9...](https://bscscan.com/tx/0x3a1bf99523474c811f58a4a764825a556ca11afba36fc76b39c753e7cd4e70b0) |
| 20 | 0x8f696A28391F3e436394E4a23773bbE972aa9e15 | 月球币 | USDT | $14000 | 0.000000012413445378 | 0.0005 | 0.00000057 | -0.00049943 | -99.89% | TP 200% | 0.000000000014223722 | 16.8 | [0xf1fc52...](https://bscscan.com/tx/0xf1fc52b14143d6b023e0bb905a4b1f4efafb0278c376142b8fc065bf2b6d23cc) | [0xf9aac2...](https://bscscan.com/tx/0xf9aac2630c04a14a013dac597241d2cd7d43fb804fadce2d3dce696d204d6c6e) |
| 22 | 0xA89D9Bb06Af0d9A6f9C4B2418F4A43448608B552 | NVO | WBNB | $28289.5 | 25.063156641605257136 | 0.0005 | 0.0004975 | -0.0000025 | -0.5% | Break-even SL | 24.937999993757930639 | 11.5 | [0xe0c043...](https://bscscan.com/tx/0xe0c043936b05949df9d231686bf400113de89001d8b1494ce048b128269af249) | [0x0354be...](https://bscscan.com/tx/0x0354bebf38f320e146060ee7d278955ec0fa3efd43826207d38255ef848c3241) |
| 24 | 0xf7BB3753E07B3970dc990BFb9419AA5C6d9802dE | 杀零狗 | USDT | $12000 | 0.000000010639046085 | 0.0005 | 0.00049502 | -0.00000498 | -1% | Break-even SL | 0.000000010533056403 | 9 | [0xb642b9...](https://bscscan.com/tx/0xb642b92abb3858db378fdb45ab3104ab6ddafdbf5e73710089b4fc7f07100b86) | [0xa0f469...](https://bscscan.com/tx/0xa0f469192d00af2e8ced6cd57d8fb764c834b1f731850fca8c3fa262d0c0c5be) |
| 26 | 0xf5B7C9fdC09445437F1eFbcAdb389e8DD80D3EAB | QQQB | USDT | $14000 | 0.000000012412136794 | 0.0005 | 0.0000002 | -0.0004998 | -99.96% | Break-even SL | 0.00000000000493815 | 12 | [0x7d7857...](https://bscscan.com/tx/0x7d7857c2a98444acf3a614c908943a358bea2b9e922591299266e21cda2150c7) | [0xfc3375...](https://bscscan.com/tx/0xfc33756637096e62864e53fa6fca5ff7b63de128a3a7f030dedb18b18cad592a) |
| 28 | 0xb897C0E2E3086D2CEa97058EB98bdE1a4d60DFE1 | QQQB | USDT | $12000 | 0.000000010639416314 | 0.0005 | 0.00049502 | -0.00000498 | -1% | Break-even SL | 0.000000010533422749 | 10.9 | [0x252da7...](https://bscscan.com/tx/0x252da7869070b291990c19d97670bc99711db9fea47b345e1ea8c1a95efa1ce4) | [0x873eab...](https://bscscan.com/tx/0x873eabbc111ef5fa20ed61b94d31a28c2f9b26c2af744cda787b93b70ac94ea0) |
| 30 | 0xA3C6bB3011f80F4C1A918af7B97B11bF95cc7a36 | 杀零狗 | USDT | $14000 | 0.000000012412568561 | 0.0005 | 0.00049502 | -0.00000498 | -1% | Break-even SL | 0.000000012288909846 | 6.9 | [0x51272d...](https://bscscan.com/tx/0x51272de13875525e2c98eed58a5dc2d6c8f9e512b21017ca1bcbd7138a6b7516) | [0x0355a9...](https://bscscan.com/tx/0x0355a9c97988c351f4833a4b466e3ae7b1e5af6ee4a8e1be236bdff333de6010) |
| 32 | 0xCddB158F63796Ef11C2eA0b8056612E2C59D6aEe | SHAZ | WBNB | $16977.9 | 150.380939849663548102 | 0.0005 | 0.0004975 | -0.0000025 | -0.5% | Break-even SL | 149.629999937579952984 | 13.6 | [0x2c5c97...](https://bscscan.com/tx/0x2c5c972dca1b879d46eb1291e8f3494ba1d110c0cf4a125d08867ad3d5bb83fd) | [0x8c838c...](https://bscscan.com/tx/0x8c838c4e405be20bad5d08f25c81506937e5476087ab3983c70369b2283cba2f) |
| 34 | 0x3bBc2E59b44Ac2E7537135714D6C53e48699D7cB | Binance | USDT | $12000 | 0.000000010639258172 | 0.0005 | 0.00000057 | -0.00049943 | -99.89% | TP 200% | 0.000000000012190759 | 16.9 | [0xbf4601...](https://bscscan.com/tx/0xbf46017df4a6e7d5630b1d3ce4886ba8e459a2acfc4c776aaf86a8270292c491) | [0xebcd55...](https://bscscan.com/tx/0xebcd55f7bf0c3601f4e2bdd095634fd1e45971ac43ddf5dd927d8199305fac06) |
| 35 | 0x9D2150d0B31f297f25656Ad6809F7281080c1C4D | LAB | USDT | $14000 | 0.00000001241236786 | 0.0005 | 0.00049502 | -0.00000498 | -1% | Break-even SL | 0.000000012288746433 | 9.6 | [0xd2029b...](https://bscscan.com/tx/0xd2029b119379f0cb17eafacd27519681d27bd227574ae8ddb124d39af598bf91) | [0xb47283...](https://bscscan.com/tx/0xb47283e8d2cd25cf08cee04c95bdbb6fe49319f0a53ba9b927409ea37f516524) |
| 37 | 0x1f07e0ea784692197AA5e9e207e09C31bA1c7CcC | BASTEROID | WBNB | $16976.1 | 15.038093984962737082 | 0.0005 | 0.0004975 | -0.0000025 | -0.5% | Break-even SL | 14.962999993758003825 | 11.5 | [0x895fba...](https://bscscan.com/tx/0x895fba4908cc36b15eecaee7d77dd871818323cde221f38fa2fded0b8dbf91f9) | [0x22378a...](https://bscscan.com/tx/0x22378ab650d894c0b9369bb36c6fbff0f602cfc17483e11e44355f73f688d713) |
| 40 | 0x3318F4F5A761C0962eB61842fc62D28B3338514F | CNET | WBNB | $6768.52 | 0.000005995487468672 | 0.0005 | 0.0000028 | -0.0004972 | -99.44% | Break-even SL | 0.000000033573762456 | 7.1 | [0x5a3237...](https://bscscan.com/tx/0x5a32375392e131072ed6763d80cd67ac947a2278199e923e99fdd6bb71f9d6f9) | [0xe25f48...](https://bscscan.com/tx/0xe25f48af53ecc18a7e5c04ee55c5bd2ff377515224b6055282ed8bd562d86727) |
| 43 | 0x14EC87Fe36dA36a2f35926d2aa763daDc2819E27 | QQQB | USDT | $12000 | 0.000000010645282725 | 0.0005 | 0.00049503 | -0.00000497 | -0.99% | Break-even SL | 0.000000010539511551 | 7.9 | [0xd5dcd3...](https://bscscan.com/tx/0xd5dcd38a68d183b0f05d59e6c40ec6954f3c8beffc0d7e00df1272c0a788f4ef) | [0xccd815...](https://bscscan.com/tx/0xccd815a4468405f697a04d4329daa00e4444ee766e9d803066689136229ad3c2) |
| 45 | 0xFaD66d1B243C80230Ad9314164B0495DcD32Fe5b | 旺旺 | USDT | $14000 | 0.000000012419743875 | 0.0005 | 0.00049504 | -0.00000496 | -0.99% | Break-even SL | 0.000000012296462262 | 7.8 | [0x5c5e50...](https://bscscan.com/tx/0x5c5e50ad660c394f2b4d1807b6bcf0e23b2ba680b2400aa1d0127895b35e44a8) | [0x6340e9...](https://bscscan.com/tx/0x6340e9c81952b3dd2a444791f2493c651abfa89907c5276a34c0e68e4ae50d4f) |
| 47 | 0x6ed803DBc75976C899bE565cB4237ebAb336B469 | Narcos. | WBNB | $65683.84 | 0.000000008426936726 | 0.0005 | 0.0004975 | -0.0000025 | -0.5% | Break-even SL | 0.000000008384855072 | 14.4 | [0x42bb2e...](https://bscscan.com/tx/0x42bb2ec3c81218041ee392d678105a9fb397bbccae6342be56d2dc40a6047cb0) | [0x056da9...](https://bscscan.com/tx/0x056da9569d73935a9fba8970b84232666fbe7008ea1e2d9323b289f57bb94b58) |

**Full transaction hashes (confirmed round-trips):**

- AIF (TP 200%) — buy: 0x1e36f0e05432511ccde0f041d58d47680ed34f6852ba8e39281d813c20b715e5 — sell: 0x0ba714f5ba2682f1bf301a31e27b7e1073651dff088acd1178025f131d2305ba
- CGN (TP 200%) — buy: 0x63584c539736485e3c1112710457bdc297240e756f59cdb96ee5bc8dea95d87d — sell: 0xfa24361fc512b260c3d56dc4b2c5a29cb6ccc6e38412391537cfff81b8b127c1
- 黑马 (Break-even SL) — buy: 0x25642a59290824d1f308f98156946b137680cf5d19af39670e5a5b5e4e22ea28 — sell: 0xeb89c692627c4990a4155b54ab54fab04ece2e4318f2be4f6aa6216be91b27c4
- CFI (TP 200%) — buy: 0x8d65bd6209e48967dadc2655d02e78e8e94afc7ee64a4c58dd5c5f670f536b90 — sell: 0x85970ced3c150b64523d4dedecace6692916e3a17b464577a80034116f128bb3
- CLBS (TP 200%) — buy: 0x77a6b42657b4110763f3f57452b7f7495c07cdd304f88a23251b632571c4a85e — sell: 0xddc736201fe0168a167a0a8375b4521d17327afa1ef7ad94e1e65c364f1ecac8
- CMND (TP 200%) — buy: 0x8b9947ae4650cfd9bae1d466f68aca43035d6cef690b94b094275da51ad163b8 — sell: 0xf7ef8a4e3799e34f0492cc9156f553bffd151dd79aba956a12534ed657a6d6b8
- CAGT (TP 200%) — buy: 0xc47c8b3a2e18545e1336620cb2de0b9a73d2cb0f4d80283a363b7742c5019f80 — sell: 0x1ee34e2af8ecb54cd1be992f763dc96459376810cabd730e3e3565c787c4cb29
- 小狐狸 (Reverted) — buy: 0x1ba5bd606514434ec3cdf53bbd9196aff052391f1ee5f327eca19671c19e8dc0 — sell: 0x3a1bf99523474c811f58a4a764825a556ca11afba36fc76b39c753e7cd4e70b0
- 月球币 (TP 200%) — buy: 0xf1fc52b14143d6b023e0bb905a4b1f4efafb0278c376142b8fc065bf2b6d23cc — sell: 0xf9aac2630c04a14a013dac597241d2cd7d43fb804fadce2d3dce696d204d6c6e
- NVO (Break-even SL) — buy: 0xe0c043936b05949df9d231686bf400113de89001d8b1494ce048b128269af249 — sell: 0x0354bebf38f320e146060ee7d278955ec0fa3efd43826207d38255ef848c3241
- 杀零狗 (Break-even SL) — buy: 0xb642b92abb3858db378fdb45ab3104ab6ddafdbf5e73710089b4fc7f07100b86 — sell: 0xa0f469192d00af2e8ced6cd57d8fb764c834b1f731850fca8c3fa262d0c0c5be
- QQQB (Break-even SL) — buy: 0x7d7857c2a98444acf3a614c908943a358bea2b9e922591299266e21cda2150c7 — sell: 0xfc33756637096e62864e53fa6fca5ff7b63de128a3a7f030dedb18b18cad592a
- QQQB (Break-even SL) — buy: 0x252da7869070b291990c19d97670bc99711db9fea47b345e1ea8c1a95efa1ce4 — sell: 0x873eabbc111ef5fa20ed61b94d31a28c2f9b26c2af744cda787b93b70ac94ea0
- 杀零狗 (Break-even SL) — buy: 0x51272de13875525e2c98eed58a5dc2d6c8f9e512b21017ca1bcbd7138a6b7516 — sell: 0x0355a9c97988c351f4833a4b466e3ae7b1e5af6ee4a8e1be236bdff333de6010
- SHAZ (Break-even SL) — buy: 0x2c5c972dca1b879d46eb1291e8f3494ba1d110c0cf4a125d08867ad3d5bb83fd — sell: 0x8c838c4e405be20bad5d08f25c81506937e5476087ab3983c70369b2283cba2f
- Binance (TP 200%) — buy: 0xbf46017df4a6e7d5630b1d3ce4886ba8e459a2acfc4c776aaf86a8270292c491 — sell: 0xebcd55f7bf0c3601f4e2bdd095634fd1e45971ac43ddf5dd927d8199305fac06
- LAB (Break-even SL) — buy: 0xd2029b119379f0cb17eafacd27519681d27bd227574ae8ddb124d39af598bf91 — sell: 0xb47283e8d2cd25cf08cee04c95bdbb6fe49319f0a53ba9b927409ea37f516524
- BASTEROID (Break-even SL) — buy: 0x895fba4908cc36b15eecaee7d77dd871818323cde221f38fa2fded0b8dbf91f9 — sell: 0x22378ab650d894c0b9369bb36c6fbff0f602cfc17483e11e44355f73f688d713
- CNET (Break-even SL) — buy: 0x5a32375392e131072ed6763d80cd67ac947a2278199e923e99fdd6bb71f9d6f9 — sell: 0xe25f48af53ecc18a7e5c04ee55c5bd2ff377515224b6055282ed8bd562d86727
- QQQB (Break-even SL) — buy: 0xd5dcd38a68d183b0f05d59e6c40ec6954f3c8beffc0d7e00df1272c0a788f4ef — sell: 0xccd815a4468405f697a04d4329daa00e4444ee766e9d803066689136229ad3c2
- 旺旺 (Break-even SL) — buy: 0x5c5e50ad660c394f2b4d1807b6bcf0e23b2ba680b2400aa1d0127895b35e44a8 — sell: 0x6340e9c81952b3dd2a444791f2493c651abfa89907c5276a34c0e68e4ae50d4f
- Narcos. (Break-even SL) — buy: 0x42bb2ec3c81218041ee392d678105a9fb397bbccae6342be56d2dc40a6047cb0 — sell: 0x056da9569d73935a9fba8970b84232666fbe7008ea1e2d9323b289f57bb94b58

## 3. Reverted / Pending Transactions (Group C)

| # | Token | Symbol | Quote | Side | Status | Liquidity (USD) | Amount BNB | Tx Hash | Reason |
|---|-------|--------|-------|------|--------|-----------------|------------|---------|--------|
| 5 | 0x60b07bC89DBe64a78dba0d7717a1Fa841540F1D3 |  | UNKNOWN | buy | pending | $8991.45 | 0.0005 | 0x946d63886bbbce09433fff606449c581dce3b3d3a3236c002830b848dd3ba563 | Pending/no receipt |
| 9 | 0x367a53283e71EECC185fcA39103656d664891F45 |  | UNKNOWN | buy | pending | $11722.55 | 0.0005 | 0x5333a7ad109687e6e66ab36539a08373116fd91a82b35776f9f62af50b8dab10 | Pending/no receipt |
| 17 | 0x8Dd53d629b0c0030654CE0b595135d76E7bD03dF |  | UNKNOWN | buy | pending | $6796.35 | 0.0005 | 0xc7ed1a0f075973c3ed261fc68d887642bd81bd56aa81d463968b24d02ffc3efb | Pending/no receipt |
| 42 | 0x9053e29201cA0F9Fd80C8C6aaa06e467FDd3c5da |  | UNKNOWN | buy | pending | $6757.2 | 0.0005 | 0x06b1361d786e3d8a37154c100739c23936952249e6dd76cc60511c5f79945f09 | Pending/no receipt |
| 19 | 0xA77b335082b87E526Cd72AE14A7C4aAe3eA9C956 | 小狐狸 | USDT | sell | pending | $12000 | 0.00049502 | 0x3a1bf99523474c811f58a4a764825a556ca11afba36fc76b39c753e7cd4e70b0 | On-chain revert |

## 4. Group Comparison Summary

| Metric | Group A: Winners (n=8) | Group B: Break-even SL (n=13) | Group C: Reverts (n=5) |
|--------|----------------------------------------|----------------------------------------|----------------------------------------|
| Avg liquidity (USD) | $8745.03 | $18515.07 | $9253.51 |
| Median liquidity (USD) | $7494.22 | $14000 | $8991.45 |
| Avg buy BNB | 0.0005 | 0.0005 | 0.000499 |
| Avg sell BNB | 0.00000239 | 0.00038181 | — |
| Avg PnL (BNB) | -0.00049761 | -0.00011819 | — |
| Avg PnL % | -99.52% | -23.64% | — |
| Avg round-trip efficiency | 0.0048 | 0.7636 | — |
| Avg time to sell (s) | 14.9 | 10.4 | — |
| Avg age at filter (s) | 1.1 | 0.7 | 0 |
| Avg holders | 0 | 0 | 0 |
| WBNB pairs | 6 | 5 | 0 |
| USDT pairs | 2 | 8 | 1 |

### Detailed Group A (Winners, TP 200%)

- **AIF** — sold for 0.000003 BNB (-99.40% vs 0.0005 BNB buy). Round-trip efficiency: 0.006. Time to sell: 14.2 s. Liquidity: $8188.01. Quote: WBNB.
- **CGN** — sold for 0.000003 BNB (-99.40% vs 0.0005 BNB buy). Round-trip efficiency: 0.006. Time to sell: 14.6 s. Liquidity: $8776.09. Quote: WBNB.
- **CFI** — sold for 0.000003 BNB (-99.40% vs 0.0005 BNB buy). Round-trip efficiency: 0.006. Time to sell: 14.3 s. Liquidity: $6755.17. Quote: WBNB.
- **CLBS** — sold for 0.000003 BNB (-99.40% vs 0.0005 BNB buy). Round-trip efficiency: 0.006. Time to sell: 15.4 s. Liquidity: $6800.44. Quote: WBNB.
- **CMND** — sold for 0.000003 BNB (-99.40% vs 0.0005 BNB buy). Round-trip efficiency: 0.006. Time to sell: 13.9 s. Liquidity: $6708.96. Quote: WBNB.
- **CAGT** — sold for 0.000003 BNB (-99.40% vs 0.0005 BNB buy). Round-trip efficiency: 0.006. Time to sell: 13.3 s. Liquidity: $6731.59. Quote: WBNB.
- **月球币** — sold for 0.00000057 BNB (-99.89% vs 0.0005 BNB buy). Round-trip efficiency: 0.0011. Time to sell: 16.8 s. Liquidity: $14000. Quote: USDT.
- **Binance** — sold for 0.00000057 BNB (-99.89% vs 0.0005 BNB buy). Round-trip efficiency: 0.0011. Time to sell: 16.9 s. Liquidity: $12000. Quote: USDT.

### Detailed Group B (Break-even SL)

- **黑马** — sold for 0.00000037 BNB (-99.93% vs 0.0005 BNB buy). Round-trip efficiency: 0.0007. Time to sell: 12.9 s. Liquidity: $14000. Quote: USDT. Estimated effective tax: 99.93%.
- **NVO** — sold for 0.0004975 BNB (-0.50% vs 0.0005 BNB buy). Round-trip efficiency: 0.995. Time to sell: 11.5 s. Liquidity: $28289.5. Quote: WBNB. Estimated effective tax: 0.5%.
- **杀零狗** — sold for 0.00049502 BNB (-1.00% vs 0.0005 BNB buy). Round-trip efficiency: 0.99. Time to sell: 9 s. Liquidity: $12000. Quote: USDT. Estimated effective tax: 1%.
- **QQQB** — sold for 0.0000002 BNB (-99.96% vs 0.0005 BNB buy). Round-trip efficiency: 0.0004. Time to sell: 12 s. Liquidity: $14000. Quote: USDT. Estimated effective tax: 99.96%.
- **QQQB** — sold for 0.00049502 BNB (-1.00% vs 0.0005 BNB buy). Round-trip efficiency: 0.99. Time to sell: 10.9 s. Liquidity: $12000. Quote: USDT. Estimated effective tax: 1%.
- **杀零狗** — sold for 0.00049502 BNB (-1.00% vs 0.0005 BNB buy). Round-trip efficiency: 0.99. Time to sell: 6.9 s. Liquidity: $14000. Quote: USDT. Estimated effective tax: 1%.
- **SHAZ** — sold for 0.0004975 BNB (-0.50% vs 0.0005 BNB buy). Round-trip efficiency: 0.995. Time to sell: 13.6 s. Liquidity: $16977.9. Quote: WBNB. Estimated effective tax: 0.5%.
- **LAB** — sold for 0.00049502 BNB (-1.00% vs 0.0005 BNB buy). Round-trip efficiency: 0.99. Time to sell: 9.6 s. Liquidity: $14000. Quote: USDT. Estimated effective tax: 1%.
- **BASTEROID** — sold for 0.0004975 BNB (-0.50% vs 0.0005 BNB buy). Round-trip efficiency: 0.995. Time to sell: 11.5 s. Liquidity: $16976.1. Quote: WBNB. Estimated effective tax: 0.5%.
- **CNET** — sold for 0.0000028 BNB (-99.44% vs 0.0005 BNB buy). Round-trip efficiency: 0.0056. Time to sell: 7.1 s. Liquidity: $6768.52. Quote: WBNB. Estimated effective tax: 99.44%.
- **QQQB** — sold for 0.00049503 BNB (-0.99% vs 0.0005 BNB buy). Round-trip efficiency: 0.9901. Time to sell: 7.9 s. Liquidity: $12000. Quote: USDT. Estimated effective tax: 0.99%.
- **旺旺** — sold for 0.00049504 BNB (-0.99% vs 0.0005 BNB buy). Round-trip efficiency: 0.9901. Time to sell: 7.8 s. Liquidity: $14000. Quote: USDT. Estimated effective tax: 0.99%.
- **Narcos.** — sold for 0.0004975 BNB (-0.50% vs 0.0005 BNB buy). Round-trip efficiency: 0.995. Time to sell: 14.4 s. Liquidity: $65683.84. Quote: WBNB. Estimated effective tax: 0.5%.

## 5. Key Insights

### 5.1 What correlated with winning (TP hit)?

- **Winners were lower-liquidity launches.** Average liquidity was $8745.03 for winners vs $18515.07 for SL exits. The 4 lowest-liquidity approved tokens that were bought (AIF, CFI, CLBS, CMND, CAGT, CNET) all hit the TP trigger.
- **Winners did NOT produce positive returns.** Despite the TP 200% flag, every winner returned < 2% of the 0.0005 BNB buy amount. Average round-trip efficiency was only 0.0048.
- **Winners were mostly WBNB pairs.** 6 of 8 winners were WBNB pairs; 2 were USDT pairs.
- **Time to sell was slightly longer for winners** (14.9 s vs 10.4 s for SL), suggesting the 3-second monitor took a few extra ticks to register and act on the spike.
- **The "TP 200%" signal is not a profit signal in this data.** The monitor detected a price ≥ 3× entry, but the actual sell output was decimated by token tax, slippage, or price manipulation.

### 5.2 What correlated with break-even SL?

- **USDT pairs overwhelmingly hit SL.** 8 of 13 SL exits were USDT pairs, while only 5 were WBNB pairs. This is the strongest single discriminator in the dataset.
- **Higher liquidity, no spike.** SL tokens had higher average liquidity ($18515.07) and never reached the 3× trigger in the monitored window.
- **Two distinct SL sub-groups:**
  - *Low-damage SL:* NVO, SHAZ, BASTEROID, Narcos., 杀零狗, QQQB, 旺旺, LAB returned ~99% of the buy (effective tax ≈ 0.5–1%).
  - *High-damage SL:* 黑马, QQQB#2, CNET returned < 1% of the buy due to severe effective tax or slippage.

### 5.3 Symbol / name patterns

- Symbols observed: `AIF`, `CGN`, `黑马`, `CFI`, `CLBS`, `CMND`, `CAGT`, `小狐狸`, `月球币`, `NVO`, `杀零狗`, `QQQB`, `QQQB`, `杀零狗`, `SHAZ`, `Binance`, `LAB`, `BASTEROID`, `CNET`, `QQQB`, `旺旺`, `Narcos.`, `小狐狸`.
- No reliable correlation between symbol length, Chinese characters, or English names and outcome. Both Chinese-character and English-symbol tokens appear in winners and losers.
- The duplicate name **QQQB** was bought twice with different contracts; both hit break-even SL, one with catastrophic effective tax. Duplicate-name risk is real.

### 5.4 Reverted transactions

- **4 buys and 1 sell were recorded as pending / reverted.** Their liquidity values were $9253.51 on average, overlapping the lower end of approved liquidity.
- The reverted sell was for **小狐狸** (USDT pair), a confirmed buy whose break-even SL sell failed on-chain. The DB still marks the position closed.
- Reverted buys consumed gas but do not show `gas_used` in the DB (status remained `pending`), likely because the receipt was not fetched before the bot stopped.

## 6. Anomalies That Need Fixing

1. **ATH vs sell price mismatch in winners.** The prior test log reported extreme price spikes (e.g., AIF +22,259%), but the DB `positions.ath_price_bnb` only stores the value at position creation, and the actual sell output is tiny. The real spike price is lost unless the monitor persists its in-memory ATH to the DB before selling.
2. **USDT-pair price conversion appears unreliable.** The monitor converts USDT reserves to BNB by dividing by the live WBNB/USDT price. 8 of the 13 SL exits were USDT pairs, and the reverted sell was also USDT. This path warrants validation.
3. **Pending reverted trades.** The DB status for the 4 reverted buys and 1 reverted sell is `pending` with `gas_used = 0`, contradicting the earlier finding that they were on-chain reverts. The executor should update `status = 'reverted'` and `gas_used` from the receipt even if the bot stops or the receipt fetch times out.
4. **No holder / top10 / rug data from BSCScan.** The `filter_results` table shows `holder_count = 0` and `top10_pct = 0` for every approved token. The BSCScan API call either failed or was not returning data, disabling the holder-concentration and rug-score safeguards for the entire test.
5. **Pre-buy tax simulation is too small.** The filter simulates a 0.001 BNB round trip and rejects tax > 15%. AIF/CGN/CFI/CLBS/CMND/CAGT all passed this check but returned < 1% of the buy amount. The simulation amount should match the real buy size (0.0005 BNB) and the primary guard should be the round-trip ratio, not the per-side tax.

## 7. Recommendations for the Next Test

### 7.1 Filter / token selection

1. **Temporarily reject USDT-paired launches.** This is the highest-confidence change: 8/13 SL exits and the only reverted sell were USDT pairs. Restrict to WBNB pairs until the conversion logic is validated.
2. **Raise the liquidity floor to at least $10,000.** The current $5,000 floor was the minimum that produced trades. Raising it to $10,000 would have filtered out the 4 reverted buys plus several low-liquidity winners that returned almost nothing.
3. **Replace the 15% tax guard with a real-size round-trip guard.** Simulate the exact 0.0005 BNB buy and immediate 100% sell. Reject tokens where the simulated BNB back is < 50% of BNB in. This would have eliminated all the 99%+ damage trades.
4. **Harden the BSCScan holder lookup.** Add retries, logging, and a fallback to on-chain holder counting so the holder/top10/rug checks actually run.
5. **Add a duplicate-name guard.** If the same symbol has already been bought in the last 5 minutes, skip the new contract to avoid QQQB-style double exposure.

### 7.2 Strategy / execution

1. **Keep the 0.0005 BNB buy size.** Gas cost is a small fraction of the total loss; the dominant issue is token selection.
2. **Persist in-memory ATH to the DB before selling.** When the monitor triggers a TP, write the current spike price to `positions.ath_price_bnb` so future analysis can compare trigger price vs execution price.
3. **Consider a partial exit instead of binary 100%.** The current model relies on a single execution price. A 50% TP at +100% and 50% TP at +300% would reduce variance from a single bad fill.
4. **Fix the reverted-status update.** Ensure the executor always writes the final receipt status (confirmed/reverted) and gas used before returning.
5. **Log actual effective tax per sell.** After a sell, compare the actual BNB balance change to the pre-sell `getAmountsOut` estimate and store the effective tax in the DB.

### 7.3 Liquidity floor

- Test a **$10,000 floor** first. It would have removed 8 of the 22 confirmed buys (the 4 reverted buys + AIF, CFI, CLBS, CMND, CAGT, CNET) and preserved most of the higher-liquidity trades. If the next test still sees low-quality tokens, raise to $15,000.

## 8. CSV Export (Machine-Readable)

```csv
id,token,symbol,quote,liquidity_usd,age_seconds,holders,top10_pct,rug_score,entry_price_bnb,buy_bnb,sell_bnb,pnl_bnb,pnl_pct,exit_type,sell_price_bnb,time_to_sell_s,round_trip_efficiency,tax_estimate_pct,buy_tx,sell_tx
1,0xD1399bE5c383f6AE0F06C42Bf4110749B8a9FC3F,AIF,WBNB,8188.0056,3,0,0,0,0.000007258645363409,0.0005,0.000003,-0.000497,-99.4,TP 200%,4.355070185e-8,14.167,0.006,null,0x1e36f0e05432511ccde0f041d58d47680ed34f6852ba8e39281d813c20b715e5,0x0ba714f5ba2682f1bf301a31e27b7e1073651dff088acd1178025f131d2305ba
3,0xF3B0545Ff4799A55D1105E068fB182C2F3498d4A,CGN,WBNB,8776.0944,0,0,0,0,7.779948622e-9,0.0005,0.000003,-0.000497,-99.4,TP 200%,4.6678436e-11,14.579,0.006,null,0x63584c539736485e3c1112710457bdc297240e756f59cdb96ee5bc8dea95d87d,0xfa24361fc512b260c3d56dc4b2c5a29cb6ccc6e38412391537cfff81b8b127c1
6,0x8dA1c053B6eDaa8144d4c144CFBd8518f79f11a0,黑马,USDT,14000,0,0,0,0,1.2412561517e-8,0.0005,3.7e-7,-0.00049963,-99.926,Break-even SL,9.307178e-12,12.865,0.00074,99.926,0x25642a59290824d1f308f98156946b137680cf5d19af39670e5a5b5e4e22ea28,0xeb89c692627c4990a4155b54ab54fab04ece2e4318f2be4f6aa6216be91b27c4
8,0xFCff67FDDf26848aEB2D27BBf37707d15dA03517,CFI,WBNB,6755.1744,3,0,0,0,5.985462406e-9,0.0005,0.000003,-0.000497,-99.4,TP 200%,3.5911809e-11,14.278,0.006,null,0x8d65bd6209e48967dadc2655d02e78e8e94afc7ee64a4c58dd5c5f670f536b90,0x85970ced3c150b64523d4dedecace6692916e3a17b464577a80034116f128bb3
11,0xC30831B3Ef404b5649b9aE193dB46EE15C34A5A2,CLBS,WBNB,6800.4352,0,0,0,0,3.01278133e-10,0.0005,0.000003,-0.000497,-99.4,TP 200%,1.80762e-12,15.36,0.006,null,0x77a6b42657b4110763f3f57452b7f7495c07cdd304f88a23251b632571c4a85e,0xddc736201fe0168a167a0a8375b4521d17327afa1ef7ad94e1e65c364f1ecac8
13,0x84120991a1227AE35fEA821A7a9aA30479f81602,CMND,WBNB,6708.9648,0,0,0,0,2.97268108e-10,0.0005,0.000003,-0.000497,-99.4,TP 200%,1.783561e-12,13.854,0.006,null,0x8b9947ae4650cfd9bae1d466f68aca43035d6cef690b94b094275da51ad163b8,0xf7ef8a4e3799e34f0492cc9156f553bffd151dd79aba956a12534ed657a6d6b8
15,0x0a3b3625b8216c8d153A25cdC611Cc531ad257c6,CAGT,WBNB,6731.592,0,0,0,0,5.9654122807e-7,0.0005,0.000003,-0.000497,-99.4,TP 200%,3.579151194e-9,13.349,0.006,null,0xc47c8b3a2e18545e1336620cb2de0b9a73d2cb0f4d80283a363b7742c5019f80,0x1ee34e2af8ecb54cd1be992f763dc96459376810cabd730e3e3565c787c4cb29
18,0xA77b335082b87E526Cd72AE14A7C4aAe3eA9C956,小狐狸,USDT,12000,0,0,0,0,1.064016825e-8,0.0005,0.00049502,-0.0005,-100,Reverted,1.0534166449e-8,10.758,0,null,0x1ba5bd606514434ec3cdf53bbd9196aff052391f1ee5f327eca19671c19e8dc0,0x3a1bf99523474c811f58a4a764825a556ca11afba36fc76b39c753e7cd4e70b0
20,0x8f696A28391F3e436394E4a23773bbE972aa9e15,月球币,USDT,14000,0,0,0,0,1.2413445378e-8,0.0005,5.7e-7,-0.00049943,-99.88600000000001,TP 200%,1.4223722e-11,16.795,0.0011400000000000002,null,0xf1fc52b14143d6b023e0bb905a4b1f4efafb0278c376142b8fc065bf2b6d23cc,0xf9aac2630c04a14a013dac597241d2cd7d43fb804fadce2d3dce696d204d6c6e
22,0xA89D9Bb06Af0d9A6f9C4B2418F4A43448608B552,NVO,WBNB,28289.5,0,0,0,0,25.063156641605257,0.0005,0.0004975,-0.0000025000000000000066,-0.5000000000000013,Break-even SL,24.93799999375793,11.542,0.995,0.5000000000000004,0xe0c043936b05949df9d231686bf400113de89001d8b1494ce048b128269af249,0x0354bebf38f320e146060ee7d278955ec0fa3efd43826207d38255ef848c3241
24,0xf7BB3753E07B3970dc990BFb9419AA5C6d9802dE,杀零狗,USDT,12000,0,0,0,0,1.0639046085e-8,0.0005,0.00049502,-0.0000049800000000000235,-0.9960000000000047,Break-even SL,1.0533056403e-8,8.959,0.9900399999999999,0.996000000000008,0xb642b92abb3858db378fdb45ab3104ab6ddafdbf5e73710089b4fc7f07100b86,0xa0f469192d00af2e8ced6cd57d8fb764c834b1f731850fca8c3fa262d0c0c5be
26,0xf5B7C9fdC09445437F1eFbcAdb389e8DD80D3EAB,QQQB,USDT,14000,0,0,0,0,1.2412136794e-8,0.0005,2e-7,-0.0004998,-99.96000000000001,Break-even SL,4.93815e-12,11.969,0.00039999999999999996,99.96000000000001,0x7d7857c2a98444acf3a614c908943a358bea2b9e922591299266e21cda2150c7,0xfc33756637096e62864e53fa6fca5ff7b63de128a3a7f030dedb18b18cad592a
28,0xb897C0E2E3086D2CEa97058EB98bdE1a4d60DFE1,QQQB,USDT,12000,3,0,0,0,1.0639416314e-8,0.0005,0.00049502,-0.0000049800000000000235,-0.9960000000000047,Break-even SL,1.0533422749e-8,10.888,0.9900399999999999,0.996000000000008,0x252da7869070b291990c19d97670bc99711db9fea47b345e1ea8c1a95efa1ce4,0x873eabbc111ef5fa20ed61b94d31a28c2f9b26c2af744cda787b93b70ac94ea0
30,0xA3C6bB3011f80F4C1A918af7B97B11bF95cc7a36,杀零狗,USDT,14000,0,0,0,0,1.2412568561e-8,0.0005,0.00049502,-0.0000049800000000000235,-0.9960000000000047,Break-even SL,1.2288909846e-8,6.942,0.9900399999999999,0.996000000000008,0x51272de13875525e2c98eed58a5dc2d6c8f9e512b21017ca1bcbd7138a6b7516,0x0355a9c97988c351f4833a4b466e3ae7b1e5af6ee4a8e1be236bdff333de6010
32,0xCddB158F63796Ef11C2eA0b8056612E2C59D6aEe,SHAZ,WBNB,16977.9,0,0,0,0,150.38093984966355,0.0005,0.0004975,-0.0000025000000000000066,-0.5000000000000013,Break-even SL,149.62999993757995,13.594,0.995,0.5000000000000004,0x2c5c972dca1b879d46eb1291e8f3494ba1d110c0cf4a125d08867ad3d5bb83fd,0x8c838c4e405be20bad5d08f25c81506937e5476087ab3983c70369b2283cba2f
34,0x3bBc2E59b44Ac2E7537135714D6C53e48699D7cB,Binance,USDT,12000,3,0,0,0,1.0639258172e-8,0.0005,5.7e-7,-0.00049943,-99.88600000000001,TP 200%,1.2190759e-11,16.852,0.0011400000000000002,null,0xbf46017df4a6e7d5630b1d3ce4886ba8e459a2acfc4c776aaf86a8270292c491,0xebcd55f7bf0c3601f4e2bdd095634fd1e45971ac43ddf5dd927d8199305fac06
35,0x9D2150d0B31f297f25656Ad6809F7281080c1C4D,LAB,USDT,14000,0,0,0,0,1.241236786e-8,0.0005,0.00049502,-0.0000049800000000000235,-0.9960000000000047,Break-even SL,1.2288746433e-8,9.569,0.9900399999999999,0.996000000000008,0xd2029b119379f0cb17eafacd27519681d27bd227574ae8ddb124d39af598bf91,0xb47283e8d2cd25cf08cee04c95bdbb6fe49319f0a53ba9b927409ea37f516524
37,0x1f07e0ea784692197AA5e9e207e09C31bA1c7CcC,BASTEROID,WBNB,16976.1,0,0,0,0,15.038093984962737,0.0005,0.0004975,-0.0000025000000000000066,-0.5000000000000013,Break-even SL,14.962999993758004,11.458,0.995,0.5000000000000004,0x895fba4908cc36b15eecaee7d77dd871818323cde221f38fa2fded0b8dbf91f9,0x22378ab650d894c0b9369bb36c6fbff0f602cfc17483e11e44355f73f688d713
40,0x3318F4F5A761C0962eB61842fc62D28B3338514F,CNET,WBNB,6768.5228,6,0,0,0,0.000005995487468672,0.0005,0.0000028,-0.0004972,-99.44000000000001,Break-even SL,3.3573762456e-8,7.12,0.0056,99.44,0x5a32375392e131072ed6763d80cd67ac947a2278199e923e99fdd6bb71f9d6f9,0xe25f48af53ecc18a7e5c04ee55c5bd2ff377515224b6055282ed8bd562d86727
43,0x14EC87Fe36dA36a2f35926d2aa763daDc2819E27,QQQB,USDT,12000,0,0,0,0,1.0645282725e-8,0.0005,0.00049503,-0.0000049699999999999744,-0.9939999999999949,Break-even SL,1.0539511551e-8,7.929,0.99006,0.9939999999999949,0xd5dcd38a68d183b0f05d59e6c40ec6954f3c8beffc0d7e00df1272c0a788f4ef,0xccd815a4468405f697a04d4329daa00e4444ee766e9d803066689136229ad3c2
45,0xFaD66d1B243C80230Ad9314164B0495DcD32Fe5b,旺旺,USDT,14000,0,0,0,0,1.2419743875e-8,0.0005,0.00049504,-0.000004960000000000034,-0.9920000000000068,Break-even SL,1.2296462262e-8,7.769,0.99008,0.992000000000004,0x5c5e50ad660c394f2b4d1807b6bcf0e23b2ba680b2400aa1d0127895b35e44a8,0x6340e9c81952b3dd2a444791f2493c651abfa89907c5276a34c0e68e4ae50d4f
47,0x6ed803DBc75976C899bE565cB4237ebAb336B469,Narcos.,WBNB,65683.84,0,0,0,0,8.426936726e-9,0.0005,0.0004975,-0.0000025000000000000066,-0.5000000000000013,Break-even SL,8.384855072e-9,14.37,0.995,0.5000000000000004,0x42bb2ec3c81218041ee392d678105a9fb397bbccae6342be56d2dc40a6047cb0,0x056da9569d73935a9fba8970b84232666fbe7008ea1e2d9323b289f57bb94b58
```

---
*Generated from local PostgreSQL at 2026-07-23T18:49:21.963Z.*
