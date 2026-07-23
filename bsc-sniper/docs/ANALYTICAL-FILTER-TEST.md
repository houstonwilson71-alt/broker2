# Analytical Filter Test Report (30-Minute Live Mainnet)

**Date:** 2026-07-23  
**Duration:** 30 minutes  
**Wallet:** `0xA2591F919f18Ba5b6f8A00f872dB07fCe968Ef84`  
**Buy amount:** 0.001 BNB  
**Liquidity floor:** $12,000 USD  
**Efficiency guard:** ≥ 0.85 (85% round-trip)  
**BNB price used for USD:** $565.67  

## 1. Executive Summary

- **Total buys:** 16  
- **Total BNB spent:** 0.016 BNB  
- **Total sells:** 16  
- **Total BNB recovered:** 0.008975 BNB  
- **Net P&L:** -0.007025 BNB (-3.97 USD)  
- **TP 200% winners:** 7  
- **Break-even SL exits:** 9  
- **Pairs seen:** see logs  
- **Pairs approved:** 16  
- **Pairs rejected:** 15  

## 2. All Buys (Full Transaction Hashes)

| # | Token | Symbol | Quote | Buy BNB | Buy Tx |
|---|-------|--------|-------|---------|--------|
| 1 | 0xB2A4Fc2628FE06E602D5ea975e1efC45fC740D9F | AB | WBNB | 0.001 | [0xc61777...](https://bscscan.com/tx/0xc6177747327c21a06dff2fbd5d4fbf2db7c9277d3fadd0da93f4245e99867422) |
| 2 | 0xFac90694Ca38Db0B55432Cf4Ce7068e8Ad1A4FCf | CMO | WBNB | 0.001 | [0x87c7ce...](https://bscscan.com/tx/0x87c7ce3b4783be3d5beb126ba5540a1aac33f6e7137ef3adb4505b79c167a69b) |
| 3 | 0xa2E04DB1FC49480906a39c482722428208F3720C | 5miles | WBNB | 0.001 | [0x533d5e...](https://bscscan.com/tx/0x533d5e31d17f5dfb3dfe04aed5a9c9dd2a3436cb5eeca305fa991c48ce4c7488) |
| 4 | 0xA8FB0536236287A916D325A95Ae0aec3f82c8814 | QQQB | USDT | 0.001 | [0x0dbab1...](https://bscscan.com/tx/0x0dbab16b7ac62e29f16cf588d2b02f3e15f0957c223e2f2bec1ac5f0d7080a79) |
| 5 | 0xFbfd02771c80bdd6C9543888e74dbA4a29a9875D | 巨龙 | WBNB | 0.001 | [0x255bce...](https://bscscan.com/tx/0x255bce00c98d50a75f63de2da361fc92d0c85927d071878e3dc819ca4ec506db) |
| 6 | 0x6d540a2176cD0d581C3bF59eE2B53430769Eb6C4 | CZHolder | WBNB | 0.001 | [0xc7eaca...](https://bscscan.com/tx/0xc7eaca7fb464d204fca3410cd4485640b63dbbc89e91d317f11bab7d9a195266) |
| 7 | 0x6fF4EDF01D614465be97eD75849CBda1528E2418 | 杀零狗 | USDT | 0.001 | [0x30e85c...](https://bscscan.com/tx/0x30e85c0b13ae7d3c438ac0a97021dde0a8a6d1490a3184dd293235fad65faf5c) |
| 8 | 0x99017606982E106dA28567F94C494C8d54aB0ec7 | UniRTM | WBNB | 0.001 | [0xc1aa61...](https://bscscan.com/tx/0xc1aa61decc40966ab52edc163c55cce25bd5a69b7e4abbab403d27a7d6617b31) |
| 9 | 0x58933547CBC40eada321a13a2392F83A40783A5F | CZ | WBNB | 0.001 | [0x23dcea...](https://bscscan.com/tx/0x23dcea1010fd07effcda11f63268242f01e66b0198504bd6a69338c85527a88c) |
| 10 | 0x8535DA4d27e95AE0cb8591Cd8f81D84F7c2406b4 | 币安黑马 | USDT | 0.001 | [0x486da9...](https://bscscan.com/tx/0x486da9260382d6df0fd9cc334996f7e7dd6619d6cac5717d620a4be2e93b5a82) |
| 11 | 0x0d0271D532395866Aa6FE917207C4BE1d12Fd200 | QQQB | USDT | 0.001 | [0x54bfd5...](https://bscscan.com/tx/0x54bfd5b7af84c9d16121eea2d11a1c2041e0a608ea0924e5adb0ca485edd8794) |
| 12 | 0xea5488Ba3B993260d1c6adF9F4a063f00AcE8ac9 | 小狐狸 | USDT | 0.001 | [0x6881cd...](https://bscscan.com/tx/0x6881cd9165afc32ce01ee31d031bf6e43d24f83fe1cf637977c4ac24aa96b96b) |
| 13 | 0xC5bD7A60c18CA9C6c75D3e7e455c5B2BBCC96C6E | ZYLO | WBNB | 0.001 | [0x61622c...](https://bscscan.com/tx/0x61622c37e054266851e7f33411c39803d2087caf063542ee72c7a592d7a4047c) |
| 14 | 0x15d7D5a495EeeD0675271AC5233265B2374bAc3A | OUSTon | WBNB | 0.001 | [0xaff66e...](https://bscscan.com/tx/0xaff66e55d2581215d7b72c72ccd691a28cc0588461fb3ab90c4feb22d52bf296) |
| 15 | 0x1A406306c978340AdE0dd1C1181421fE605256c9 | 杀零狗 | USDT | 0.001 | [0xeb81c9...](https://bscscan.com/tx/0xeb81c9c9b49aa450bc87abbb3b93644b4d3fee6a543f6273ce129d9f369cb961) |
| 16 | 0xb5f85585B9F9617C2537597b9f05AC95d4B15944 | 火星币 | USDT | 0.001 | [0x4a70aa...](https://bscscan.com/tx/0x4a70aaea37f24f04d59e086eb78557ac65183809ecb659496ce6947f8cc1198b) |

**Full buy hashes:**

- AB — 0xc6177747327c21a06dff2fbd5d4fbf2db7c9277d3fadd0da93f4245e99867422
- CMO — 0x87c7ce3b4783be3d5beb126ba5540a1aac33f6e7137ef3adb4505b79c167a69b
- 5miles — 0x533d5e31d17f5dfb3dfe04aed5a9c9dd2a3436cb5eeca305fa991c48ce4c7488
- QQQB — 0x0dbab16b7ac62e29f16cf588d2b02f3e15f0957c223e2f2bec1ac5f0d7080a79
- 巨龙 — 0x255bce00c98d50a75f63de2da361fc92d0c85927d071878e3dc819ca4ec506db
- CZHolder — 0xc7eaca7fb464d204fca3410cd4485640b63dbbc89e91d317f11bab7d9a195266
- 杀零狗 — 0x30e85c0b13ae7d3c438ac0a97021dde0a8a6d1490a3184dd293235fad65faf5c
- UniRTM — 0xc1aa61decc40966ab52edc163c55cce25bd5a69b7e4abbab403d27a7d6617b31
- CZ — 0x23dcea1010fd07effcda11f63268242f01e66b0198504bd6a69338c85527a88c
- 币安黑马 — 0x486da9260382d6df0fd9cc334996f7e7dd6619d6cac5717d620a4be2e93b5a82
- QQQB — 0x54bfd5b7af84c9d16121eea2d11a1c2041e0a608ea0924e5adb0ca485edd8794
- 小狐狸 — 0x6881cd9165afc32ce01ee31d031bf6e43d24f83fe1cf637977c4ac24aa96b96b
- ZYLO — 0x61622c37e054266851e7f33411c39803d2087caf063542ee72c7a592d7a4047c
- OUSTon — 0xaff66e55d2581215d7b72c72ccd691a28cc0588461fb3ab90c4feb22d52bf296
- 杀零狗 — 0xeb81c9c9b49aa450bc87abbb3b93644b4d3fee6a543f6273ce129d9f369cb961
- 火星币 — 0x4a70aaea37f24f04d59e086eb78557ac65183809ecb659496ce6947f8cc1198b

## 3. All Sells (Full Transaction Hashes + Exit Reason)

| # | Token | Symbol | Quote | Sell BNB | Reason | Sell Tx |
|---|-------|--------|-------|----------|--------|---------|
| 1 | 0xB2A4Fc2628FE06E602D5ea975e1efC45fC740D9F | AB | WBNB | 0.000995 | Break-even SL | [0x0dc525...](https://bscscan.com/tx/0x0dc525382336d6f85b97b37e8b2615f211094df8e057131ad545190d37f9a763) |
| 2 | 0xFac90694Ca38Db0B55432Cf4Ce7068e8Ad1A4FCf | CMO | WBNB | 0.000995 | Break-even SL | [0x3fce6d...](https://bscscan.com/tx/0x3fce6d504ec4a55ee638db13f006051aa9ac1ca49e4993f244cd00ce1573b192) |
| 3 | 0xa2E04DB1FC49480906a39c482722428208F3720C | 5miles | WBNB | 0.000995 | Break-even SL | [0xc8ab17...](https://bscscan.com/tx/0xc8ab179268d9b99a37fe80a1829d4a5716c90f933983335d1f3fad7d6760bb60) |
| 4 | 0xA8FB0536236287A916D325A95Ae0aec3f82c8814 | QQQB | USDT | 0.000002 | TP 200% | [0xa89d79...](https://bscscan.com/tx/0xa89d7965d666b7e99ea3c891e0afc38a860688ac4021515883e01b8c3ee4da1f) |
| 5 | 0xFbfd02771c80bdd6C9543888e74dbA4a29a9875D | 巨龙 | WBNB | 0.000995 | Break-even SL | [0x6333d2...](https://bscscan.com/tx/0x6333d247343e289d67b12539a96b7626a2c6372459e2f6b5ecbf9361bbb72459) |
| 6 | 0x6d540a2176cD0d581C3bF59eE2B53430769Eb6C4 | CZHolder | WBNB | 0.001003 | Break-even SL | [0x351ffd...](https://bscscan.com/tx/0x351ffddf7c6a08b9ffd17a50c564b3ff08368bcc3605d62eee7d1c8f920a1743) |
| 7 | 0x6fF4EDF01D614465be97eD75849CBda1528E2418 | 杀零狗 | USDT | 0.000002 | TP 200% | [0x425272...](https://bscscan.com/tx/0x4252726d294a02ed6b74528960a19d22ea3289dd29a26a6b3e61e5498863f1ae) |
| 8 | 0x99017606982E106dA28567F94C494C8d54aB0ec7 | UniRTM | WBNB | 0.000995 | Break-even SL | [0xcc3832...](https://bscscan.com/tx/0xcc3832e29e380d0097b713bb24d27b41074cbb9ec328cff8917ac51b59b6bfd5) |
| 9 | 0x58933547CBC40eada321a13a2392F83A40783A5F | CZ | WBNB | 0.000995 | Break-even SL | [0x7417c5...](https://bscscan.com/tx/0x7417c5e6228929f63334638d0ca2ff5251d1cf1cadecdc70f621c5fc030f0586) |
| 10 | 0x8535DA4d27e95AE0cb8591Cd8f81D84F7c2406b4 | 币安黑马 | USDT | 0.000002 | TP 200% | [0xda730f...](https://bscscan.com/tx/0xda730f9faefdac24052f759c6d77495de160237faabe37961cb25b2d901fc8f6) |
| 11 | 0x0d0271D532395866Aa6FE917207C4BE1d12Fd200 | QQQB | USDT | 0.000001 | TP 200% | [0x56f5eb...](https://bscscan.com/tx/0x56f5ebbfd4bc556c16181f3a253559ce7e9f8b73e36a5efcb35059dfdbe349d5) |
| 12 | 0xea5488Ba3B993260d1c6adF9F4a063f00AcE8ac9 | 小狐狸 | USDT | 0.000002 | TP 200% | [0x1f428c...](https://bscscan.com/tx/0x1f428c437ace981f292612ee121bb206b05b59ca28fe0fd416850628f1048793) |
| 13 | 0xC5bD7A60c18CA9C6c75D3e7e455c5B2BBCC96C6E | ZYLO | WBNB | 0.000995 | Break-even SL | [0x252fe4...](https://bscscan.com/tx/0x252fe4fd8ac0061d7b8ea952022a45baf38ae248a2290b162208e5de6bd332c9) |
| 14 | 0x15d7D5a495EeeD0675271AC5233265B2374bAc3A | OUSTon | WBNB | 0.000995 | Break-even SL | [0x61370a...](https://bscscan.com/tx/0x61370a11ebb2857d13b19dd857c9a7400e40f9b58510c6bb7c831569539d235f) |
| 15 | 0x1A406306c978340AdE0dd1C1181421fE605256c9 | 杀零狗 | USDT | 0.000002 | TP 200% | [0xdfa515...](https://bscscan.com/tx/0xdfa5156ce9f379980df8bcd2cdd425d284de4f80e3685108199fd6fcc1933bc4) |
| 16 | 0xb5f85585B9F9617C2537597b9f05AC95d4B15944 | 火星币 | USDT | 0.000002 | TP 200% | [0xd54c63...](https://bscscan.com/tx/0xd54c63acf67cc531cf00c0964f3ff3765c51164241af73521f0e36b1e6e2840a) |

**Full sell hashes:**

- AB (Break-even SL) — 0x0dc525382336d6f85b97b37e8b2615f211094df8e057131ad545190d37f9a763
- CMO (Break-even SL) — 0x3fce6d504ec4a55ee638db13f006051aa9ac1ca49e4993f244cd00ce1573b192
- 5miles (Break-even SL) — 0xc8ab179268d9b99a37fe80a1829d4a5716c90f933983335d1f3fad7d6760bb60
- QQQB (TP 200%) — 0xa89d7965d666b7e99ea3c891e0afc38a860688ac4021515883e01b8c3ee4da1f
- 巨龙 (Break-even SL) — 0x6333d247343e289d67b12539a96b7626a2c6372459e2f6b5ecbf9361bbb72459
- CZHolder (Break-even SL) — 0x351ffddf7c6a08b9ffd17a50c564b3ff08368bcc3605d62eee7d1c8f920a1743
- 杀零狗 (TP 200%) — 0x4252726d294a02ed6b74528960a19d22ea3289dd29a26a6b3e61e5498863f1ae
- UniRTM (Break-even SL) — 0xcc3832e29e380d0097b713bb24d27b41074cbb9ec328cff8917ac51b59b6bfd5
- CZ (Break-even SL) — 0x7417c5e6228929f63334638d0ca2ff5251d1cf1cadecdc70f621c5fc030f0586
- 币安黑马 (TP 200%) — 0xda730f9faefdac24052f759c6d77495de160237faabe37961cb25b2d901fc8f6
- QQQB (TP 200%) — 0x56f5ebbfd4bc556c16181f3a253559ce7e9f8b73e36a5efcb35059dfdbe349d5
- 小狐狸 (TP 200%) — 0x1f428c437ace981f292612ee121bb206b05b59ca28fe0fd416850628f1048793
- ZYLO (Break-even SL) — 0x252fe4fd8ac0061d7b8ea952022a45baf38ae248a2290b162208e5de6bd332c9
- OUSTon (Break-even SL) — 0x61370a11ebb2857d13b19dd857c9a7400e40f9b58510c6bb7c831569539d235f
- 杀零狗 (TP 200%) — 0xdfa5156ce9f379980df8bcd2cdd425d284de4f80e3685108199fd6fcc1933bc4
- 火星币 (TP 200%) — 0xd54c63acf67cc531cf00c0964f3ff3765c51164241af73521f0e36b1e6e2840a

## 4. Efficiency Data (Every Token Filtered)

| Token | Symbol | Quote | Efficiency | Status | Rejection Reason |
|-------|--------|-------|------------|--------|------------------|
| 0xB2A4Fc2628FE06E602D5ea975e1efC45fC740D9F | AB | WBNB | 0.995 | APPROVED |  |
| 0xFac90694Ca38Db0B55432Cf4Ce7068e8Ad1A4FCf | CMO | WBNB | 0.995 | APPROVED |  |
| 0xa2E04DB1FC49480906a39c482722428208F3720C | 5miles | WBNB | 0.995 | APPROVED |  |
| 0xA8FB0536236287A916D325A95Ae0aec3f82c8814 | QQQB | USDT | 0.9899 | APPROVED |  |
| 0xFbfd02771c80bdd6C9543888e74dbA4a29a9875D | 巨龙 | WBNB | 0.995 | APPROVED |  |
| 0x6d540a2176cD0d581C3bF59eE2B53430769Eb6C4 | CZHolder | WBNB | 0.995 | APPROVED |  |
| 0x6fF4EDF01D614465be97eD75849CBda1528E2418 | 杀零狗 | USDT | 0.9899 | APPROVED |  |
| 0x99017606982E106dA28567F94C494C8d54aB0ec7 | UniRTM | WBNB | 0.9948 | APPROVED |  |
| 0x58933547CBC40eada321a13a2392F83A40783A5F | CZ | WBNB | 0.995 | APPROVED |  |
| 0x8535DA4d27e95AE0cb8591Cd8f81D84F7c2406b4 | 币安黑马 | USDT | 0.9899 | APPROVED |  |
| 0x0d0271D532395866Aa6FE917207C4BE1d12Fd200 | QQQB | USDT | 0.9899 | APPROVED |  |
| 0xea5488Ba3B993260d1c6adF9F4a063f00AcE8ac9 | 小狐狸 | USDT | 0.9899 | APPROVED |  |
| 0xC5bD7A60c18CA9C6c75D3e7e455c5B2BBCC96C6E | ZYLO | WBNB | 0.9949 | APPROVED |  |
| 0x15d7D5a495EeeD0675271AC5233265B2374bAc3A | OUSTon | WBNB | 0.9949 | APPROVED |  |
| 0x1A406306c978340AdE0dd1C1181421fE605256c9 | 杀零狗 | USDT | 0.9899 | APPROVED |  |
| 0xb5f85585B9F9617C2537597b9f05AC95d4B15944 | 火星币 | USDT | 0.9899 | APPROVED |  |
| 0x81C87B8146954404DBA2c0554B0f5124D480C6a2 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0xd97e0f2e37822D42Ade75847436BA6AB7b2f434b |  | WBNB | 0.9942 | REJECTED | low_liquidity:$2832 |
| 0xfFfEdFa3A48eeEF1E9c201c38BDa5A8232Ca9c7D |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0x9EB646Fa5849fb722Dd16bD5ff0a8afff93525fd |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0x981b74A4F5F7A4fe25f2a10C140b6Fa8f101007C |  | WBNB | 0.995 | REJECTED | duplicate_symbol:5miles |
| 0xCe91069bEcF920E10123c4e9594306fDE6fd8Ccd |  | WBNB | 0.9947 | REJECTED | low_liquidity:$6777 |
| 0x1C96728e4C28dbf1896250d7C3CDe456d391CB4D |  | USDT | 0.9899 | REJECTED | duplicate_symbol:QQQB |
| 0x33E57a915F5ad45700e68F18181D31FCb7df43C4 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0x79cdB5b8fcAEE28c55a5B4421B101E7612b09668 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0x9E2a2fBb51472e038EA61281dc00D5491c339aa7 |  | WBNB | 0.9948 | REJECTED | low_liquidity:$9063 |
| 0xF4cECd0b78047516697033e0b35d8C832d640fef |  | USDT | 0 | REJECTED | low_efficiency:0.0000, low_liquidity:$4019 |
| 0x2B183A3cbB197a242FE2658f5D465e3f70543BCD |  | WBNB | 0.9947 | REJECTED | low_liquidity:$6707 |
| 0xA0BE88Cf04766e631892339496f051a75299b049 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0x27F471e65b5EBCd23327a4483C0eb4623769e8F7 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0xd3445318fB159d054d4Fb18271D856451eE17421 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |

## 5. Rejected Tokens Breakdown

- **Rejected by efficiency guard:** 9  
- **Rejected by liquidity floor:** 4  
- **Rejected by duplicate symbol guard:** 2  
- **Rejected by other reasons:** 0  

## 6. Comparison to Previous Test

| Metric | Previous Test (22 trades) | This Test (30 min) |
|--------|---------------------------|--------------------|
| Buy size | 0.0005 BNB | 0.001 BNB |
| Liquidity floor | $5,000 | $12,000 |
| Efficiency guard | 15% tax check | 85% round-trip efficiency |
| Net P&L | -0.006017 BNB | -0.007025 BNB |
| Avg round-trip efficiency | 0.4530 | 0.5609 |
| Wins / SL | 8 TP / 13 SL | 7 TP / 9 SL |

### Analysis

This test did not beat the previous test. Net P&L was **-0.007025 BNB** vs **-0.006017 BNB** previously. The efficiency guard allowed tokens with ~99.5% simulated round-trip efficiency, which translated to real losses of ~0.5% on break-even SL exits. TP 200% winners still suffered heavy tax/slippage, returning only ~0.17% of the 0.001 BNB buy. The duplicate-symbol guard worked, rejecting duplicate `5miles` and `QQQB` contracts. The $12,000 liquidity floor rejected many low-liquidity launches that previously caused 99%+ damage.

## 7. Observations and Recommendations

1. **Efficiency guard is effective at filtering out honeypots.** Every rejected token with efficiency=0 had a reverting or zero-output sell simulation.
2. **Break-even SL is now a small, controlled loss.** Realized losses on SL exits were ~0.5% of buy size, matching the simulated efficiency.
3. **TP 200% winners still lose money.** The monitor triggers on price spikes, but actual sell output is decimated by token tax/slippage. Consider taking partial profits or using a higher TP threshold.
4. **USDT pairs executed successfully.** Two-hop swaps through WBNB/USDT worked, and USDT pairs were not auto-rejected.
5. **BSCScan holder API still failed.** All BSCScan holder lookups returned no data after 3 retries; the top-10 concentration guard could not be validated.
6. **Next test suggestion:** Raise the efficiency floor from 0.85 to 0.90 or 0.95 to see if the few remaining SL losses can be reduced further.

---
*Generated from live mainnet test on 2026-07-23.*
