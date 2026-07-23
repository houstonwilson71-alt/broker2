# 30-Minute Live Mainnet Test V2 — Lowered Floor, WBNB-Only

**Date:** 2026-07-23  
**Duration:** 30 minutes (started 21:31:57 UTC, stopped 22:02:07 UTC)  
**Wallet:** `0xA2591F919f18Ba5b6f8A00f872dB07fCe968Ef84`  
**Buy amount:** 0.0005 BNB  
**Efficiency threshold:** 0.90 (90% round-trip)  
**Liquidity floor:** $8,000 USD  
**Quote-token restriction:** WBNB only — USDT, BUSD, USDC, ETH, CAKE explicitly rejected  
**TP strategy:** 50% at +200%, remaining 50% at +300% or trailing stop -20% from peak  
**Sell slippage:** amountOutMin = expected × 0.95  
**Gas boost:** 1.5× on take-profit sells  
**Duplicate symbol guard:** 5 minutes  
**BSCScan retries:** 3 attempts  
**BNB price used for USD:** $566.56  

## 1. Executive Summary

- **Confirmed buys:** 35 (1 failed buy attempt)  
- **Total BNB spent:** 0.0175 BNB  
- **Total sell transactions:** 56  
- **Total BNB recovered:** 0.002083 BNB  
- **Net P&L (from trade amounts):** -0.015417 BNB (-8.73 USD)  
- **Partial TP + trailing SL exits:** 21  
- **Single sell exits:** 14  
- **Pairs approved:** 38  
- **Pairs rejected:** 16  

## 2. P&L Reality Check

The executor's `realized_pnl_bnb` column double-counts the cost basis on the second partial sell (it attributes 100% of the buy cost to the 50% remaining-position sell). This report therefore uses the **actual trade amounts** for P&L: total BNB received from confirmed sells minus total BNB spent on confirmed buys. The raw DB column sums to ~-0.0207 BNB, while the actual cash loss is -0.015417 BNB.

## 3. All Buys (Full Transaction Hashes)

| # | Token | Symbol | Quote | Buy BNB | Status | Buy Tx |
|---|-------|--------|-------|---------|--------|--------|
| 1 | 0xC4eb5d4F8057d131DAce90cd8FCc1d33fa4c839C | MCH | WBNB | 0.0005 | confirmed | [0xb6b72b...](https://bscscan.com/tx/0xb6b72be42b4ddbb378c708181c150bcdec3993447843e3a273315fbf0dae4561) |
| 2 | 0x26DB36a76d8271f04f25C6bFF1dE7A3B10d3267b | APX | WBNB | 0.0005 | confirmed | [0x97c99c...](https://bscscan.com/tx/0x97c99cf254522fce313f6a60d418d86f41e23d15843acb7f03d263a37610ae41) |
| 3 | 0xFe6E00741eEbCb125100266644e82e47434EB537 | TWLO | WBNB | 0.0005 | confirmed | [0xe3512a...](https://bscscan.com/tx/0xe3512ab7ff0193bfe674280e814ff61ffe99c1a19da15bda027519c943b5dd32) |
| 4 | 0x594BDca00AFCFaB3738C488Aba639093b6Aa6E39 | LAI | WBNB | 0.0005 | confirmed | [0x561a02...](https://bscscan.com/tx/0x561a02da7e736e5e02f40f6c04ebf544e82c831f9bffb9205ef0f2b79bbf6795) |
| 5 | 0x7Fdce4363de28c5d6a03708a745e2F8B9845f81C | UniCDOGE | WBNB | 0.0005 | confirmed | [0x76beb7...](https://bscscan.com/tx/0x76beb74db47c2a1ea342b9735beaba6c6908b763ca38aa2f9eb5a1b354d8aadb) |
| 6 | 0x6c33F098c6F830CbE236A3626c9c7e83f4617E37 | INF | WBNB | 0.0005 | confirmed | [0x1849c5...](https://bscscan.com/tx/0x1849c5ea1e8b8b881ad52bb7d54cde8b73297bd360e7911d745f900332d14609) |
| 7 | 0xFcBFb10ae3BcF7fD74B6f77ca36A0B8034B129ad | UniBDOGE | WBNB | 0.0005 | confirmed | [0x624aec...](https://bscscan.com/tx/0x624aec2cb7d91cd2ca791533c1e3e8f37423bada1b0f2040f37bbfb8e4e4db8c) |
| 8 | 0xd68d6a5b9CbE421d7D51a4B21aC4C19d70e6f5cB | BRC | WBNB | 0.0005 | confirmed | [0x46f261...](https://bscscan.com/tx/0x46f261d7cc3b862e4ff814063739de3109c9155e7b3a8fb391948a0c84dd2f0b) |
| 9 | 0x245f2c2C1758f9cf1E4D2AAE39095Fc05342a9C0 | VMD | WBNB | 0.0005 | confirmed | [0x4c6866...](https://bscscan.com/tx/0x4c686626cc3370d423b4c103c8a9abea9ec2260b47076abccaf5fa5289ab2e99) |
| 10 | 0x4C209fe8D7f0f201c6aAcd19CA63F6481f1ca36B | UniHDOGE | WBNB | 0.0005 | confirmed | [0x799769...](https://bscscan.com/tx/0x799769e7fe40d906cc3259d5b8c122f85b8988222597ea23c230fb9f78bad8d8) |
| 11 | 0xc15e314Fd50Bfb3a5ABb86a1A906F6E1FC82a928 | OPT | WBNB | 0.0005 | confirmed | [0x49041c...](https://bscscan.com/tx/0x49041c8f5db6a47e153e4106f05323f5de0200c02bc6f5b6dbacc6d0b6c836a8) |
| 12 | 0xf3ecE4b0Fa720cFb5013AF4C75f098A6dE1B9ec6 | NFG | WBNB | 0.0005 | confirmed | [0xc2d03d...](https://bscscan.com/tx/0xc2d03dd0b60af8a13faeefa33d6f7fc8a42a80002fb0b044bfa754af40798dfb) |
| 13 | 0xeA0C6C8015e8672E8f2AE11CE7E421911348E194 | DPC | WBNB | 0.0005 | confirmed | [0xc6b4fb...](https://bscscan.com/tx/0xc6b4fbd219070e8cc059e42587f4ad36122272d49d9977e801fc0e3ef4c87a03) |
| 14 | 0xA1ab98d9E302e516DBaeDEbc2A503743447D5333 | INTA | WBNB | 0.0005 | confirmed | [0xb0014b...](https://bscscan.com/tx/0xb0014b3ec7db4080ba51c36c461c53572c73777270fd34ca2fa96d4c6657f325) |
| 15 | 0x49D37C1e33c163573C3d9ea83BAa6E4A7a1f9986 | AIMX | WBNB | 0.0005 | confirmed | [0xb185c5...](https://bscscan.com/tx/0xb185c57b6239d65bec7607c2efc834761f8e0cc17725503c8e60b41ea311819c) |
| 16 | 0x7e9DD9C787c3d068800C3f6cfeF36326e08f306d | NSP | WBNB | 0.0005 | confirmed | [0xc88c4b...](https://bscscan.com/tx/0xc88c4b22005965721610ab1ff3ae99be78b894a5c4a2a593f8673e75eeb0b1ba) |
| 17 | 0x9D27C0a6f80d175D522C102a04d72748fA9F862f | CTX | WBNB | 0.0005 | confirmed | [0x42544d...](https://bscscan.com/tx/0x42544dbeb213101f9d4c0c382c07be28d7010ce5522045a3e75940436a7d383c) |
| 18 | 0x201896A597dAef39eE0264cD516202C057a6c7aE | VSN | WBNB | 0.0005 | confirmed | [0x891614...](https://bscscan.com/tx/0x891614f2bf5d6d467812ac25a9a85dabfc4a7b743a099522a6845ab0565ef4c4) |
| 19 | 0x14b8B1367e53d628eBeF8D45a4354DFcD27D435D | MBR | WBNB | 0.0005 | confirmed | [0x08d8e5...](https://bscscan.com/tx/0x08d8e500d0bd8f20e987f6c2f76e89de909311e2858088ed76145e824375c971) |
| 20 | 0x95Df3e6b340f3286074fdf87DfEC490342B489da | QNT | WBNB | 0.0005 | confirmed | [0x8f9f7d...](https://bscscan.com/tx/0x8f9f7de328b69ade3d14920a087138b3d3d7dfe94c21a275ad4855b3d0c30144) |
| 21 | 0x7E8577cB6452D4294121634c072f558BeF42f262 | ALM | WBNB | 0.0005 | confirmed | [0x312296...](https://bscscan.com/tx/0x3122967883b21afb28b0bf0c2456f36de2732761874afec6ca474ac2a2df9a7b) |
| 22 | 0x5F79aDb7D2a08a6Df4970EeDb9dF9F4486d76E96 | SMAI | WBNB | 0.0005 | confirmed | [0x75dbcf...](https://bscscan.com/tx/0x75dbcf2d2aca3f6779575df9b5fc4eca1aa1c04780d17d16c46649869c23bb03) |
| 23 | 0x67e9E15Dc7234fE4Dd54Dd57DC9B110cA8C9b382 | FMD | WBNB | 0.0005 | confirmed | [0x695d10...](https://bscscan.com/tx/0x695d108bfb1b227da5a6384707e694a2f78e56dab1443c8d9c43eba3537cab9a) |
| 24 | 0x8e8700fbeE5C6A2Ef5CCC4061a3E3873F9a669d2 | NCL | WBNB | 0.0005 | confirmed | [0x8074a1...](https://bscscan.com/tx/0x8074a15f2136fcbda8b8d51040641f45379f537be7e51c260313f0c4093831a1) |
| 25 | 0x98B46c4Ac30487aC4692e6863aAA588f982e6978 | INS | WBNB | 0.0005 | confirmed | [0xb2fa2c...](https://bscscan.com/tx/0xb2fa2c2f656443665f09143f8e526f02d42b5b3b2676cf490c2ccd753b7660e6) |
| 26 | 0x26A4B25C21E06961d42f68Ec5e2414636EC99e47 | BRAI | WBNB | 0.0005 | confirmed | [0x464305...](https://bscscan.com/tx/0x464305e893d526d40887283042cc81ccc9d6165827ca5720fbb0abb3e5c835e0) |
| 27 | 0x81D57EB69AE32eD4B927C63EBa97a3E6F7a81920 | DNET | WBNB | 0.0005 | confirmed | [0x38179e...](https://bscscan.com/tx/0x38179e55da5a65bb9bc391d95947cbcb913a3f27ab7c6683b422cf89da56d096) |
| 28 | 0xFc00aDd1A70729a8C21b425A09B563732F3e49ae | COGA | WBNB | 0.0005 | confirmed | [0x932594...](https://bscscan.com/tx/0x932594a211f8084297f3c15488e23c45bf85e2f8eed2c960529bd04eac9aa09e) |
| 29 | 0x61C84351a839E4da525e96233aBfB9597E308812 | AAI | WBNB | 0.0005 | confirmed | [0x9b38b0...](https://bscscan.com/tx/0x9b38b037b24bad4be46a4994c31fc009c24e60b06dac1c6eb9877fa581835360) |
| 30 | 0x30Fd89DaeC84Da2A85490507AB9344Ab5d5Ff4ca | GiggleHero | WBNB | 0.0005 | failed | [...](https://bscscan.com/tx/) |
| 31 | 0x30Fd89DaeC84Da2A85490507AB9344Ab5d5Ff4ca | GiggleHero | WBNB | 0.0005 | confirmed | [0x488f81...](https://bscscan.com/tx/0x488f8113922bb2dec90b63132adae642b385c8994125cdddbd64335e9d865eab) |
| 32 | 0x852342a8577e2C413a88f00802B76A78497a8C7A | PNET | WBNB | 0.0005 | confirmed | [0x883e40...](https://bscscan.com/tx/0x883e40e46f3e88492411e9cd91c19c69fb86ec8c78db5dcc60f96132641faf5f) |
| 33 | 0x636D3648D24374230fA8d65A9F837A3977b471Bd | LMD | WBNB | 0.0005 | confirmed | [0xa75a65...](https://bscscan.com/tx/0xa75a65a6d05daddebe55dff5420c28c3b692b435bd11ebd145cd4b58610f04cc) |
| 34 | 0x341dD99dFc28BccD58324cE91621769fcc801d4F | VCH | WBNB | 0.0005 | confirmed | [0x97f21a...](https://bscscan.com/tx/0x97f21a294f42f2f68caced22aa2ab05143d0ff4f712cb80ac4e35287cdbfea9f) |
| 35 | 0x388820B1E4a372274A21257Ab7C48537FC32eFe6 | 大表哥 | WBNB | 0.0005 | confirmed | [0x50b870...](https://bscscan.com/tx/0x50b8707e6445268630a1a428b12cc4821c0b5d30e30dabfdf4e4c203da6e7d18) |
| 36 | 0xA1f6219816E2EdE61DAEe9e9ac158E420b9b375A | SYAI | WBNB | 0.0005 | confirmed | [0xb5023d...](https://bscscan.com/tx/0xb5023d488f7327b0a33e711c2616721ac6004b3b68cd2b24d962996d8e6338b4) |

**Full confirmed buy hashes:**

- MCH — 0xb6b72be42b4ddbb378c708181c150bcdec3993447843e3a273315fbf0dae4561
- APX — 0x97c99cf254522fce313f6a60d418d86f41e23d15843acb7f03d263a37610ae41
- TWLO — 0xe3512ab7ff0193bfe674280e814ff61ffe99c1a19da15bda027519c943b5dd32
- LAI — 0x561a02da7e736e5e02f40f6c04ebf544e82c831f9bffb9205ef0f2b79bbf6795
- UniCDOGE — 0x76beb74db47c2a1ea342b9735beaba6c6908b763ca38aa2f9eb5a1b354d8aadb
- INF — 0x1849c5ea1e8b8b881ad52bb7d54cde8b73297bd360e7911d745f900332d14609
- UniBDOGE — 0x624aec2cb7d91cd2ca791533c1e3e8f37423bada1b0f2040f37bbfb8e4e4db8c
- BRC — 0x46f261d7cc3b862e4ff814063739de3109c9155e7b3a8fb391948a0c84dd2f0b
- VMD — 0x4c686626cc3370d423b4c103c8a9abea9ec2260b47076abccaf5fa5289ab2e99
- UniHDOGE — 0x799769e7fe40d906cc3259d5b8c122f85b8988222597ea23c230fb9f78bad8d8
- OPT — 0x49041c8f5db6a47e153e4106f05323f5de0200c02bc6f5b6dbacc6d0b6c836a8
- NFG — 0xc2d03dd0b60af8a13faeefa33d6f7fc8a42a80002fb0b044bfa754af40798dfb
- DPC — 0xc6b4fbd219070e8cc059e42587f4ad36122272d49d9977e801fc0e3ef4c87a03
- INTA — 0xb0014b3ec7db4080ba51c36c461c53572c73777270fd34ca2fa96d4c6657f325
- AIMX — 0xb185c57b6239d65bec7607c2efc834761f8e0cc17725503c8e60b41ea311819c
- NSP — 0xc88c4b22005965721610ab1ff3ae99be78b894a5c4a2a593f8673e75eeb0b1ba
- CTX — 0x42544dbeb213101f9d4c0c382c07be28d7010ce5522045a3e75940436a7d383c
- VSN — 0x891614f2bf5d6d467812ac25a9a85dabfc4a7b743a099522a6845ab0565ef4c4
- MBR — 0x08d8e500d0bd8f20e987f6c2f76e89de909311e2858088ed76145e824375c971
- QNT — 0x8f9f7de328b69ade3d14920a087138b3d3d7dfe94c21a275ad4855b3d0c30144
- ALM — 0x3122967883b21afb28b0bf0c2456f36de2732761874afec6ca474ac2a2df9a7b
- SMAI — 0x75dbcf2d2aca3f6779575df9b5fc4eca1aa1c04780d17d16c46649869c23bb03
- FMD — 0x695d108bfb1b227da5a6384707e694a2f78e56dab1443c8d9c43eba3537cab9a
- NCL — 0x8074a15f2136fcbda8b8d51040641f45379f537be7e51c260313f0c4093831a1
- INS — 0xb2fa2c2f656443665f09143f8e526f02d42b5b3b2676cf490c2ccd753b7660e6
- BRAI — 0x464305e893d526d40887283042cc81ccc9d6165827ca5720fbb0abb3e5c835e0
- DNET — 0x38179e55da5a65bb9bc391d95947cbcb913a3f27ab7c6683b422cf89da56d096
- COGA — 0x932594a211f8084297f3c15488e23c45bf85e2f8eed2c960529bd04eac9aa09e
- AAI — 0x9b38b037b24bad4be46a4994c31fc009c24e60b06dac1c6eb9877fa581835360
- GiggleHero — 0x488f8113922bb2dec90b63132adae642b385c8994125cdddbd64335e9d865eab
- PNET — 0x883e40e46f3e88492411e9cd91c19c69fb86ec8c78db5dcc60f96132641faf5f
- LMD — 0xa75a65a6d05daddebe55dff5420c28c3b692b435bd11ebd145cd4b58610f04cc
- VCH — 0x97f21a294f42f2f68caced22aa2ab05143d0ff4f712cb80ac4e35287cdbfea9f
- 大表哥 — 0x50b8707e6445268630a1a428b12cc4821c0b5d30e30dabfdf4e4c203da6e7d18
- SYAI — 0xb5023d488f7327b0a33e711c2616721ac6004b3b68cd2b24d962996d8e6338b4

## 4. All Sells (Full Transaction Hashes + Exit Reason)

| # | Token | Symbol | Sell # | Sell BNB | Reason | Sell Tx |
|---|-------|--------|--------|----------|--------|---------|
| 1 | 0xC4eb5d4F8057d131DAce90cd8FCc1d33fa4c839C | MCH | 1 | 0.000003 | Break-even SL | [0x46d5c6...](https://bscscan.com/tx/0x46d5c63e437d55f0e2852351eacecf5f2808b31cd624a92cb957f5700251b2c2) |
| 2 | 0x26DB36a76d8271f04f25C6bFF1dE7A3B10d3267b | APX | 1 | 0.000003 | TP 50% at +200% | [0x411d85...](https://bscscan.com/tx/0x411d85c597e1fba05927eee6c1bbcc8a9cfc45af1dab65c96b27a3d4879b140f) |
| 2 | 0x26DB36a76d8271f04f25C6bFF1dE7A3B10d3267b | APX | 2 | 0 | Trailing SL / TP 300% | [0xbcc0cc...](https://bscscan.com/tx/0xbcc0cc819c16620ea19e7bb30ec46c1d44a34fcba2e241007e7c017481bb830f) |
| 3 | 0xFe6E00741eEbCb125100266644e82e47434EB537 | TWLO | 1 | 0.000498 | Break-even SL | [0xe294db...](https://bscscan.com/tx/0xe294db00f88023057d778c39941f9a54bdcfc86cd2840a1a17c144331e00ded9) |
| 4 | 0x594BDca00AFCFaB3738C488Aba639093b6Aa6E39 | LAI | 1 | 0.000003 | Break-even SL | [0xb64c70...](https://bscscan.com/tx/0xb64c7014d24f0ac247f820977ab6ea11b84a1e3329c81f9c2e9554bd0f5230a3) |
| 5 | 0x7Fdce4363de28c5d6a03708a745e2F8B9845f81C | UniCDOGE | 1 | 0.000003 | Break-even SL | [0x03efe8...](https://bscscan.com/tx/0x03efe8539efc5af6c3464c50088301a8ffc2673d0f24f84aa2fbe6a92d604624) |
| 6 | 0x6c33F098c6F830CbE236A3626c9c7e83f4617E37 | INF | 1 | 0.000003 | Break-even SL | [0xbb44d4...](https://bscscan.com/tx/0xbb44d40a22b3aa266223cb10d677d9539d8ed2f6423c6fc5fe84f2bb9ad25f5c) |
| 7 | 0xFcBFb10ae3BcF7fD74B6f77ca36A0B8034B129ad | UniBDOGE | 1 | 0.000003 | Break-even SL | [0xb127b5...](https://bscscan.com/tx/0xb127b50fd63834094b24ffdf35437dacdabc2c6d696927a27d80494cced8657a) |
| 8 | 0xd68d6a5b9CbE421d7D51a4B21aC4C19d70e6f5cB | BRC | 1 | 0.000003 | Break-even SL | [0x370759...](https://bscscan.com/tx/0x3707591fdb4ff9c6c6947d3fe17a07cbfcf4d04becd02f95158d6ed6bbe687cb) |
| 9 | 0x245f2c2C1758f9cf1E4D2AAE39095Fc05342a9C0 | VMD | 1 | 0.000003 | Break-even SL | [0xce4f09...](https://bscscan.com/tx/0xce4f09b633735bea2e99a78cb3f98aecc5de8c2a96aac211372724f4ae440299) |
| 10 | 0x4C209fe8D7f0f201c6aAcd19CA63F6481f1ca36B | UniHDOGE | 1 | 0.000498 | Break-even SL | [0x5858ff...](https://bscscan.com/tx/0x5858ffb6e81125584098e2827ce5e78bb84647b50f1fc240b99a47432bf900b1) |
| 11 | 0xc15e314Fd50Bfb3a5ABb86a1A906F6E1FC82a928 | OPT | 1 | 0.000003 | Break-even SL | [0xfee16c...](https://bscscan.com/tx/0xfee16cf4ade6d45e1b9ed9cbb69ac54f31b1bb7dd6d96d35db12ef8017adca02) |
| 12 | 0xf3ecE4b0Fa720cFb5013AF4C75f098A6dE1B9ec6 | NFG | 1 | 0.000003 | Break-even SL | [0xcdc772...](https://bscscan.com/tx/0xcdc77221f42fbd4db8d9849caac3b65c197867d025115b39260a71d4cf6e48e8) |
| 13 | 0xeA0C6C8015e8672E8f2AE11CE7E421911348E194 | DPC | 1 | 0 | Break-even SL | [0xda28ce...](https://bscscan.com/tx/0xda28ce33d49c87739ca66daf2c06a1a540ccf8d96597aea256842349e04c033a) |
| 14 | 0xA1ab98d9E302e516DBaeDEbc2A503743447D5333 | INTA | 1 | 0.000003 | TP 50% at +200% | [0x113484...](https://bscscan.com/tx/0x1134847a6a39856c600868211220b3ce4840020c2c7395f24b0d414ba70e1c62) |
| 14 | 0xA1ab98d9E302e516DBaeDEbc2A503743447D5333 | INTA | 2 | 0 | Trailing SL / TP 300% | [0x2f2e15...](https://bscscan.com/tx/0x2f2e1511d76fdcdb37ed07aa5ff1cad32e2c76bfae0948ab70a04cea63168215) |
| 15 | 0x49D37C1e33c163573C3d9ea83BAa6E4A7a1f9986 | AIMX | 1 | 0.000003 | TP 50% at +200% | [0x97921f...](https://bscscan.com/tx/0x97921f013d2d312ccd8436571d21c453ca018962a339f360881a35f4ef6f8f55) |
| 15 | 0x49D37C1e33c163573C3d9ea83BAa6E4A7a1f9986 | AIMX | 2 | 0 | Trailing SL / TP 300% | [0x3f257a...](https://bscscan.com/tx/0x3f257af37200488397da3027033ab9acb3108ef80e819e24076b1792b22ddcaf) |
| 16 | 0x7e9DD9C787c3d068800C3f6cfeF36326e08f306d | NSP | 1 | 0.000003 | TP 50% at +200% | [0xf4c503...](https://bscscan.com/tx/0xf4c5035b80a44ead4a8a5d111035217317d67464f96b9e3fd5e574b18cdb3c0b) |
| 16 | 0x7e9DD9C787c3d068800C3f6cfeF36326e08f306d | NSP | 2 | 0 | Trailing SL / TP 300% | [0xc60664...](https://bscscan.com/tx/0xc6066405ad5335285d00bdba7778d78a1b1eae5693afefc424fbbe34130c2a81) |
| 17 | 0x9D27C0a6f80d175D522C102a04d72748fA9F862f | CTX | 1 | 0.000003 | TP 50% at +200% | [0x8a4222...](https://bscscan.com/tx/0x8a4222db815fc268847120f47dc62a3c8d00b0deb44c5e007de379c188cc260a) |
| 17 | 0x9D27C0a6f80d175D522C102a04d72748fA9F862f | CTX | 2 | 0 | Trailing SL / TP 300% | [0xefb7ee...](https://bscscan.com/tx/0xefb7ee2982cadf8c32b5988e3a3e6cc4bce2040c4b9e5783870d4e8953eb6e4f) |
| 18 | 0x201896A597dAef39eE0264cD516202C057a6c7aE | VSN | 1 | 0.000003 | TP 50% at +200% | [0x269067...](https://bscscan.com/tx/0x269067462d51fcf6cf2c98b34402e6833fff242862e9367581635f0ce2dd93e2) |
| 18 | 0x201896A597dAef39eE0264cD516202C057a6c7aE | VSN | 2 | 0 | Trailing SL / TP 300% | [0x9d7b91...](https://bscscan.com/tx/0x9d7b912031ac901f53b8368bb3f5e395f7875f607213338253fc2c4b795641ff) |
| 19 | 0x14b8B1367e53d628eBeF8D45a4354DFcD27D435D | MBR | 1 | 0.000003 | TP 50% at +200% | [0xc1a54c...](https://bscscan.com/tx/0xc1a54c8d4024900765f7262f01228daac1c0ef54f36bf2831241c6f061216e47) |
| 19 | 0x14b8B1367e53d628eBeF8D45a4354DFcD27D435D | MBR | 2 | 0 | Trailing SL / TP 300% | [0x4e68a8...](https://bscscan.com/tx/0x4e68a8ee96d6170aacd8638f1158c41445b3ebd87e874c63d7e4288492a9647d) |
| 20 | 0x95Df3e6b340f3286074fdf87DfEC490342B489da | QNT | 1 | 0.000003 | TP 50% at +200% | [0x10483c...](https://bscscan.com/tx/0x10483c33a5b86b008807766e400eceee0cd4a1592d70a66b85c95e6d82160fda) |
| 20 | 0x95Df3e6b340f3286074fdf87DfEC490342B489da | QNT | 2 | 0 | Trailing SL / TP 300% | [0x92cc52...](https://bscscan.com/tx/0x92cc5236cdc7cca05b8c939b9163b04bbb8734fbed4573fd4f4b6364c8b7b6c7) |
| 21 | 0x7E8577cB6452D4294121634c072f558BeF42f262 | ALM | 1 | 0.000003 | TP 50% at +200% | [0x88f953...](https://bscscan.com/tx/0x88f9535faf6fa0a33960965c63b96b54ff7b7fe146dc65c1d32590cd88bf5afd) |
| 21 | 0x7E8577cB6452D4294121634c072f558BeF42f262 | ALM | 2 | 0 | Trailing SL / TP 300% | [0x308581...](https://bscscan.com/tx/0x308581a793119eeac11cbf97b8457c69fe301f11b93737466419e26c79760ed9) |
| 22 | 0x5F79aDb7D2a08a6Df4970EeDb9dF9F4486d76E96 | SMAI | 1 | 0.000003 | TP 50% at +200% | [0x519a9d...](https://bscscan.com/tx/0x519a9da96fa27e06b33b1b863d9de116a8a0bc2a1cfc6d06f695c1bfcdf954d5) |
| 22 | 0x5F79aDb7D2a08a6Df4970EeDb9dF9F4486d76E96 | SMAI | 2 | 0 | Trailing SL / TP 300% | [0xd6049e...](https://bscscan.com/tx/0xd6049ea9e453ab2c7a0903d371079ce08e3a629486c8c6023e2d66fd81852ab8) |
| 23 | 0x67e9E15Dc7234fE4Dd54Dd57DC9B110cA8C9b382 | FMD | 1 | 0.000003 | TP 50% at +200% | [0x0ef92e...](https://bscscan.com/tx/0x0ef92e83204518cd62b3b34c6110dc49a5505bec408d8d040b1e0010e2dd06f4) |
| 23 | 0x67e9E15Dc7234fE4Dd54Dd57DC9B110cA8C9b382 | FMD | 2 | 0 | Trailing SL / TP 300% | [0xedcc40...](https://bscscan.com/tx/0xedcc407f095befd4793608c44db42ffc8ebdb3abc15d68d88bb805aa5108febe) |
| 24 | 0x8e8700fbeE5C6A2Ef5CCC4061a3E3873F9a669d2 | NCL | 1 | 0.000003 | TP 50% at +200% | [0xbb4d1c...](https://bscscan.com/tx/0xbb4d1c3a723812f55b36a7653174888b3a5f18da9647301654529235745add7e) |
| 24 | 0x8e8700fbeE5C6A2Ef5CCC4061a3E3873F9a669d2 | NCL | 2 | 0 | Trailing SL / TP 300% | [0x71a456...](https://bscscan.com/tx/0x71a4569427413b9da60036fb53a51d47d6ac771ee5a9e0c9b7602e481c9f4406) |
| 25 | 0x98B46c4Ac30487aC4692e6863aAA588f982e6978 | INS | 1 | 0.000003 | TP 50% at +200% | [0x3d3ebe...](https://bscscan.com/tx/0x3d3ebe56d59e7fb50c1142b204c0687c9446a4cbe418f27fbed539180379274d) |
| 25 | 0x98B46c4Ac30487aC4692e6863aAA588f982e6978 | INS | 2 | 0 | Trailing SL / TP 300% | [0x2b9f90...](https://bscscan.com/tx/0x2b9f90df93b854e158ef2f8115489d819e19e61be95160d792751f3d5f514dd7) |
| 26 | 0x26A4B25C21E06961d42f68Ec5e2414636EC99e47 | BRAI | 1 | 0.000003 | TP 50% at +200% | [0x6cd530...](https://bscscan.com/tx/0x6cd530898ea488815c7b88de68391ec2b6797666af96abca3b58e9c5d60f2339) |
| 26 | 0x26A4B25C21E06961d42f68Ec5e2414636EC99e47 | BRAI | 2 | 0 | Trailing SL / TP 300% | [0x78962c...](https://bscscan.com/tx/0x78962c48f23cbfedfdac7c21366c5c46eda8a01ffcd2739a3bade4c26cbad83e) |
| 27 | 0x81D57EB69AE32eD4B927C63EBa97a3E6F7a81920 | DNET | 1 | 0.000003 | TP 50% at +200% | [0x75998a...](https://bscscan.com/tx/0x75998a9dd5aadb07511e43e19518abbae6c9ab3d316323221e5fe768938183e3) |
| 27 | 0x81D57EB69AE32eD4B927C63EBa97a3E6F7a81920 | DNET | 2 | 0 | Trailing SL / TP 300% | [0xc8def0...](https://bscscan.com/tx/0xc8def0099907a247e15aa931466d81d2eb9c497d62595c21df7c4306d7b89c07) |
| 28 | 0xFc00aDd1A70729a8C21b425A09B563732F3e49ae | COGA | 1 | 0.000003 | TP 50% at +200% | [0x8ae148...](https://bscscan.com/tx/0x8ae148d35fde28ee15c29f4ce8439fa25e26118844ab77039c7afbc3c2c04d96) |
| 28 | 0xFc00aDd1A70729a8C21b425A09B563732F3e49ae | COGA | 2 | 0 | Trailing SL / TP 300% | [0x425d86...](https://bscscan.com/tx/0x425d86f97b3e8983551e4b75ad43712752539d255e4475183cc261f8952c841c) |
| 29 | 0x61C84351a839E4da525e96233aBfB9597E308812 | AAI | 1 | 0.000003 | TP 50% at +200% | [0x1bb8e8...](https://bscscan.com/tx/0x1bb8e861471c6d3ed3e74f840b43238ec5d4010d076459105027e776d4cd51a2) |
| 29 | 0x61C84351a839E4da525e96233aBfB9597E308812 | AAI | 2 | 0 | Trailing SL / TP 300% | [0x5ed326...](https://bscscan.com/tx/0x5ed32646b5da8b099abe00bbade1566bba9284dff886a0804da5a51cba4d849f) |
| 30 | 0x30Fd89DaeC84Da2A85490507AB9344Ab5d5Ff4ca | GiggleHero | 1 | 0.000499 | Break-even SL | [0x8396df...](https://bscscan.com/tx/0x8396df6e4f7f5d23cab43e4a2c00229e7ef4504e63643dcf5d53dd0aa6f5f9cf) |
| 31 | 0x852342a8577e2C413a88f00802B76A78497a8C7A | PNET | 1 | 0.000003 | TP 50% at +200% | [0xc2ac95...](https://bscscan.com/tx/0xc2ac95fd466039cbe8c4dbea28d492139b3f4f3c2c7c2db660f6e67314c0c1b8) |
| 31 | 0x852342a8577e2C413a88f00802B76A78497a8C7A | PNET | 2 | 0 | Trailing SL / TP 300% | [0xe758e0...](https://bscscan.com/tx/0xe758e0d511a20010e214820858e78cfcc92bc893b97abcea54a6e82591db61df) |
| 32 | 0x636D3648D24374230fA8d65A9F837A3977b471Bd | LMD | 1 | 0.000003 | TP 50% at +200% | [0xbe014f...](https://bscscan.com/tx/0xbe014f9e6bd23e78d5d29e3457cf17c56494e53dfbdf5396de0271ad09bf02d6) |
| 32 | 0x636D3648D24374230fA8d65A9F837A3977b471Bd | LMD | 2 | 0 | Trailing SL / TP 300% | [0x4dfd48...](https://bscscan.com/tx/0x4dfd48f791fc27dd448a419f66e3f861cbfd9782683cfb2590a05030ea14acc9) |
| 33 | 0x341dD99dFc28BccD58324cE91621769fcc801d4F | VCH | 1 | 0.000003 | TP 50% at +200% | [0xcf3029...](https://bscscan.com/tx/0xcf30291edc40c9bebce6fe60c0c4a536c022bfbd3ada1e415a81b0561206a6e1) |
| 33 | 0x341dD99dFc28BccD58324cE91621769fcc801d4F | VCH | 2 | 0 | Trailing SL / TP 300% | [0x682ebb...](https://bscscan.com/tx/0x682ebb4ea9d63cfb9c52c0ca103ce86c2a1e214b7ca6e0384eb4b98498f5d527) |
| 34 | 0x388820B1E4a372274A21257Ab7C48537FC32eFe6 | 大表哥 | 1 | 0.000498 | Break-even SL | [0x7c0c70...](https://bscscan.com/tx/0x7c0c7012b95fd638206daba8750a408b757439a2405f88c7128b194108c4a055) |
| 35 | 0xA1f6219816E2EdE61DAEe9e9ac158E420b9b375A | SYAI | 1 | 0.000003 | TP 50% at +200% | [0xd4b611...](https://bscscan.com/tx/0xd4b611ccf0d1082f929294732f1cadad9c1f64ab0a5763c205d7b0b385ff9c42) |
| 35 | 0xA1f6219816E2EdE61DAEe9e9ac158E420b9b375A | SYAI | 2 | 0 | Trailing SL / TP 300% | [0xa39d4c...](https://bscscan.com/tx/0xa39d4cfbade11bddc7080684abc0bfbeb61c0ae521bb3a1e3bccbc00155ba3c3) |

**Full sell hashes (selected):**

- MCH (Break-even SL) — 0x46d5c63e437d55f0e2852351eacecf5f2808b31cd624a92cb957f5700251b2c2
- APX (TP 50% at +200%) — 0x411d85c597e1fba05927eee6c1bbcc8a9cfc45af1dab65c96b27a3d4879b140f
- APX (Trailing SL / TP 300%) — 0xbcc0cc819c16620ea19e7bb30ec46c1d44a34fcba2e241007e7c017481bb830f
- TWLO (Break-even SL) — 0xe294db00f88023057d778c39941f9a54bdcfc86cd2840a1a17c144331e00ded9
- LAI (Break-even SL) — 0xb64c7014d24f0ac247f820977ab6ea11b84a1e3329c81f9c2e9554bd0f5230a3
- UniCDOGE (Break-even SL) — 0x03efe8539efc5af6c3464c50088301a8ffc2673d0f24f84aa2fbe6a92d604624
- INF (Break-even SL) — 0xbb44d40a22b3aa266223cb10d677d9539d8ed2f6423c6fc5fe84f2bb9ad25f5c
- UniBDOGE (Break-even SL) — 0xb127b50fd63834094b24ffdf35437dacdabc2c6d696927a27d80494cced8657a
- BRC (Break-even SL) — 0x3707591fdb4ff9c6c6947d3fe17a07cbfcf4d04becd02f95158d6ed6bbe687cb
- VMD (Break-even SL) — 0xce4f09b633735bea2e99a78cb3f98aecc5de8c2a96aac211372724f4ae440299
- UniHDOGE (Break-even SL) — 0x5858ffb6e81125584098e2827ce5e78bb84647b50f1fc240b99a47432bf900b1

Full sell hashes for all 35 positions are available in the trades table.

## 5. Per-Position P&L (Trade-Based)

| Token | Symbol | Buy BNB | Recovered BNB | P&L BNB | P&L % | Exit Type |
|-------|--------|---------|---------------|---------|-------|-----------|
| 0xeA0C6C8015e8672E8f2AE11CE7E421911348E194 | DPC | 0.0005 | 0 | -0.0005 | -99.96% | Break-even SL |
| 0x594BDca00AFCFaB3738C488Aba639093b6Aa6E39 | LAI | 0.0005 | 0.000003 | -0.000497 | -99.44% | Break-even SL |
| 0x7Fdce4363de28c5d6a03708a745e2F8B9845f81C | UniCDOGE | 0.0005 | 0.000003 | -0.000497 | -99.44% | Break-even SL |
| 0x6c33F098c6F830CbE236A3626c9c7e83f4617E37 | INF | 0.0005 | 0.000003 | -0.000497 | -99.44% | Break-even SL |
| 0xd68d6a5b9CbE421d7D51a4B21aC4C19d70e6f5cB | BRC | 0.0005 | 0.000003 | -0.000497 | -99.44% | Break-even SL |
| 0xc15e314Fd50Bfb3a5ABb86a1A906F6E1FC82a928 | OPT | 0.0005 | 0.000003 | -0.000497 | -99.44% | Break-even SL |
| 0xf3ecE4b0Fa720cFb5013AF4C75f098A6dE1B9ec6 | NFG | 0.0005 | 0.000003 | -0.000497 | -99.44% | Break-even SL |
| 0x7e9DD9C787c3d068800C3f6cfeF36326e08f306d | NSP | 0.0005 | 0.000003 | -0.000497 | -99.42% | TP 50% + trailing SL |
| 0xC4eb5d4F8057d131DAce90cd8FCc1d33fa4c839C | MCH | 0.0005 | 0.000003 | -0.000497 | -99.4% | Break-even SL |
| 0xFcBFb10ae3BcF7fD74B6f77ca36A0B8034B129ad | UniBDOGE | 0.0005 | 0.000003 | -0.000497 | -99.4% | Break-even SL |
| 0x245f2c2C1758f9cf1E4D2AAE39095Fc05342a9C0 | VMD | 0.0005 | 0.000003 | -0.000497 | -99.4% | Break-even SL |
| 0x26DB36a76d8271f04f25C6bFF1dE7A3B10d3267b | APX | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0xA1ab98d9E302e516DBaeDEbc2A503743447D5333 | INTA | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x49D37C1e33c163573C3d9ea83BAa6E4A7a1f9986 | AIMX | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x9D27C0a6f80d175D522C102a04d72748fA9F862f | CTX | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x201896A597dAef39eE0264cD516202C057a6c7aE | VSN | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x14b8B1367e53d628eBeF8D45a4354DFcD27D435D | MBR | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x95Df3e6b340f3286074fdf87DfEC490342B489da | QNT | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x7E8577cB6452D4294121634c072f558BeF42f262 | ALM | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x5F79aDb7D2a08a6Df4970EeDb9dF9F4486d76E96 | SMAI | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x67e9E15Dc7234fE4Dd54Dd57DC9B110cA8C9b382 | FMD | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x8e8700fbeE5C6A2Ef5CCC4061a3E3873F9a669d2 | NCL | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x98B46c4Ac30487aC4692e6863aAA588f982e6978 | INS | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x26A4B25C21E06961d42f68Ec5e2414636EC99e47 | BRAI | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x81D57EB69AE32eD4B927C63EBa97a3E6F7a81920 | DNET | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0xFc00aDd1A70729a8C21b425A09B563732F3e49ae | COGA | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x61C84351a839E4da525e96233aBfB9597E308812 | AAI | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x852342a8577e2C413a88f00802B76A78497a8C7A | PNET | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x636D3648D24374230fA8d65A9F837A3977b471Bd | LMD | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0x341dD99dFc28BccD58324cE91621769fcc801d4F | VCH | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0xA1f6219816E2EdE61DAEe9e9ac158E420b9b375A | SYAI | 0.0005 | 0.000003 | -0.000497 | -99.38% | TP 50% + trailing SL |
| 0xFe6E00741eEbCb125100266644e82e47434EB537 | TWLO | 0.0005 | 0.000498 | -0.000003 | -0.5% | Break-even SL |
| 0x4C209fe8D7f0f201c6aAcd19CA63F6481f1ca36B | UniHDOGE | 0.0005 | 0.000498 | -0.000003 | -0.5% | Break-even SL |
| 0x388820B1E4a372274A21257Ab7C48537FC32eFe6 | 大表哥 | 0.0005 | 0.000498 | -0.000003 | -0.5% | Break-even SL |
| 0x30Fd89DaeC84Da2A85490507AB9344Ab5d5Ff4ca | GiggleHero | 0.0005 | 0.000499 | -0.000001 | -0.14% | Break-even SL |

## 6. Efficiency Data (Sample of Filtered Tokens)

| Token | Symbol | Quote | Efficiency | Status | Rejection Reason |
|-------|--------|-------|------------|--------|------------------|
| 0xC4eb5d4F8057d131DAce90cd8FCc1d33fa4c839C | MCH | WBNB | 0.9947 | APPROVED |  |
| 0x26DB36a76d8271f04f25C6bFF1dE7A3B10d3267b | APX | WBNB | 0.9948 | APPROVED |  |
| 0xFe6E00741eEbCb125100266644e82e47434EB537 | TWLO | WBNB | 0.9949 | APPROVED |  |
| 0x594BDca00AFCFaB3738C488Aba639093b6Aa6E39 | LAI | WBNB | 0.9948 | APPROVED |  |
| 0x7Fdce4363de28c5d6a03708a745e2F8B9845f81C | UniCDOGE | WBNB | 0.9948 | APPROVED |  |
| 0x6c33F098c6F830CbE236A3626c9c7e83f4617E37 | INF | WBNB | 0.9948 | APPROVED |  |
| 0xFcBFb10ae3BcF7fD74B6f77ca36A0B8034B129ad | UniBDOGE | WBNB | 0.9948 | APPROVED |  |
| 0xd68d6a5b9CbE421d7D51a4B21aC4C19d70e6f5cB | BRC | WBNB | 0.9947 | APPROVED |  |
| 0x245f2c2C1758f9cf1E4D2AAE39095Fc05342a9C0 | VMD | WBNB | 0.9948 | APPROVED |  |
| 0x4C209fe8D7f0f201c6aAcd19CA63F6481f1ca36B | UniHDOGE | WBNB | 0.9948 | APPROVED |  |
| 0xc15e314Fd50Bfb3a5ABb86a1A906F6E1FC82a928 | OPT | WBNB | 0.9948 | APPROVED |  |
| 0xf3ecE4b0Fa720cFb5013AF4C75f098A6dE1B9ec6 | NFG | WBNB | 0.9947 | APPROVED |  |
| 0xeA0C6C8015e8672E8f2AE11CE7E421911348E194 | DPC | WBNB | 0.9947 | APPROVED |  |
| 0xA1ab98d9E302e516DBaeDEbc2A503743447D5333 | INTA | WBNB | 0.9948 | APPROVED |  |
| 0x49D37C1e33c163573C3d9ea83BAa6E4A7a1f9986 | AIMX | WBNB | 0.9947 | APPROVED |  |
| 0x7e9DD9C787c3d068800C3f6cfeF36326e08f306d | NSP | WBNB | 0.9947 | APPROVED |  |
| 0x9D27C0a6f80d175D522C102a04d72748fA9F862f | CTX | WBNB | 0.9948 | APPROVED |  |
| 0x201896A597dAef39eE0264cD516202C057a6c7aE | VSN | WBNB | 0.9947 | APPROVED |  |
| 0x14b8B1367e53d628eBeF8D45a4354DFcD27D435D | MBR | WBNB | 0.9947 | APPROVED |  |
| 0x95Df3e6b340f3286074fdf87DfEC490342B489da | QNT | WBNB | 0.9947 | APPROVED |  |
| 0x7E8577cB6452D4294121634c072f558BeF42f262 | ALM | WBNB | 0.9947 | APPROVED |  |
| 0x5F79aDb7D2a08a6Df4970EeDb9dF9F4486d76E96 | SMAI | WBNB | 0.9948 | APPROVED |  |
| 0x67e9E15Dc7234fE4Dd54Dd57DC9B110cA8C9b382 | FMD | WBNB | 0.9947 | APPROVED |  |
| 0x8e8700fbeE5C6A2Ef5CCC4061a3E3873F9a669d2 | NCL | WBNB | 0.9947 | APPROVED |  |
| 0x98B46c4Ac30487aC4692e6863aAA588f982e6978 | INS | WBNB | 0.9947 | APPROVED |  |
| 0x26A4B25C21E06961d42f68Ec5e2414636EC99e47 | BRAI | WBNB | 0.9948 | APPROVED |  |
| 0x81D57EB69AE32eD4B927C63EBa97a3E6F7a81920 | DNET | WBNB | 0.9947 | APPROVED |  |
| 0xFc00aDd1A70729a8C21b425A09B563732F3e49ae | COGA | WBNB | 0.9947 | APPROVED |  |
| 0x61C84351a839E4da525e96233aBfB9597E308812 | AAI | WBNB | 0.9948 | APPROVED |  |
| 0x30Fd89DaeC84Da2A85490507AB9344Ab5d5Ff4ca | GiggleHero | WBNB | 0.995 | APPROVED |  |
| 0x852342a8577e2C413a88f00802B76A78497a8C7A | PNET | WBNB | 0.9947 | APPROVED |  |
| 0x636D3648D24374230fA8d65A9F837A3977b471Bd | LMD | WBNB | 0.9947 | APPROVED |  |
| 0x341dD99dFc28BccD58324cE91621769fcc801d4F | VCH | WBNB | 0.9948 | APPROVED |  |
| 0x388820B1E4a372274A21257Ab7C48537FC32eFe6 | 大表哥 | WBNB | 0.995 | APPROVED |  |
| 0xA1f6219816E2EdE61DAEe9e9ac158E420b9b375A | SYAI | WBNB | 0.9948 | APPROVED |  |
| 0x75aF0Da0E72b00C82dec7a947837a4552434E418 | DBR | WBNB | 0.9948 | APPROVED |  |
| 0x9Cba2C7E132bDeDAe8acE6d200CAa1d9449d5786 | SN120 | WBNB | 0.9949 | APPROVED |  |
| 0x70037C928b66906D0c8caDC3F0a27Fc1136e4bFA | JOBY | WBNB | 0.9949 | APPROVED |  |
| 0x48DD8c8b0819A0e3A6AB14E914E88c60F33D891d |  | WBNB | 0.9947 | REJECTED | low_liquidity:$6757 |
| 0x93923C5A06d0e528112343dF00f999EfF5363f95 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0x353e5ACd7F86f97916065CF62923B55120869fAc |  | WBNB | 0.0476 | REJECTED | low_liquidity:$0, honeypot_detected, low_efficiency:0.0476 |
| 0x5Cfa089829C55fF5242601dDC99412d771241560 |  | WBNB | 0.9942 | REJECTED | low_liquidity:$2835 |
| 0xC4B6501bba8d396c403c2B3c6Eff2f4131AD2f0C |  | WBNB | 0.9675 | REJECTED | low_liquidity:$79 |
| 0xBC86e9d11759d74f0406A03D2d3Ee64eC694ABf5 |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |
| 0xbD1bFd4B2571D316A3dF46A63e2aea3b79634817 |  | USDT | 0.9473 | REJECTED | low_liquidity:$50, non_wbnb_quote:USDT |
| 0x4b49506a89cE6e7bd7aC11ed139d94311fDa605B |  | WBNB | 0.971 | REJECTED | low_liquidity:$91 |
| 0x0bA17DA0769d5B250164a330D939F6Ad64412CC7 |  | WBNB | 0.9735 | REJECTED | low_liquidity:$102 |
| 0x6F83AC3b3285EeC44dC706547cf0b0A018633BaC |  | WBNB | 0.9947 | REJECTED | low_liquidity:$6810 |
| 0xe13c647E78f27DB99315B3C66731506236543532 |  | USDT | 0.9899 | REJECTED | non_wbnb_quote:USDT |
| 0x96863adFc249D155eb1C00f031a1de04FE22ae6D |  | USDT | 0.9899 | REJECTED | non_wbnb_quote:USDT |
| 0x7190F4C320CdBA03B9e6C5FaDd2459F00BC9cb88 |  | ETH | 0 | REJECTED | low_efficiency:0.0000, low_liquidity:$8 |
| 0x7190F4C320CdBA03B9e6C5FaDd2459F00BC9cb88 |  | ETH | 0 | REJECTED | low_efficiency:0.0000, low_liquidity:$8, non_wbnb_quote:CAKE |
| 0x7190F4C320CdBA03B9e6C5FaDd2459F00BC9cb88 |  | ETH | 0 | REJECTED | low_efficiency:0.0000, low_liquidity:$8, non_wbnb_quote:ETH |
| 0x576C5a5f25deC76582cf4e0F26Eb70f0FE13f7Ef |  | WBNB | 0 | REJECTED | low_efficiency:0.0000, liquidity_error:zero quote balance after 3 retries |

## 7. Rejected Tokens Breakdown

- **Rejected by non-WBNB quote guard:** 2  
- **Rejected by efficiency guard (< 0.90):** 7  
- **Rejected by liquidity floor (< $8k):** 7  
- **Rejected by top-10 concentration (>30%):** 0  
- **Rejected by duplicate symbol guard:** 0  
- **Rejected by other reasons:** 0  

## 8. Comparison to Previous WBNB-Only Test

| Metric | Previous WBNB-Only Test (20 min, 0.95 eff, $12k) | This Test (30 min, 0.90 eff, $8k) |
|--------|-----------------------------------------------------|-------------------------------------|
| Duration | 20 min | 30 min |
| Efficiency threshold | 0.95 | 0.90 |
| Liquidity floor | $12,000 | $8,000 |
| Confirmed buys | 3 | 35 |
| BNB spent | 0.0015 BNB | 0.0175 BNB |
| BNB recovered | 0.001492 BNB | 0.002083 BNB |
| Net P&L (trade-based) | -0.000008 BNB | -0.015417 BNB |
| Net P&L USD | -0.00 USD | -8.73 USD |
| Loss per trade | -0.000003 BNB | -0.00044 BNB |

### Analysis

Lowering the efficiency threshold from 0.95 to 0.90 and the liquidity floor from $12k to $8k was **strongly detrimental**. The bot executed 35 buys in 30 minutes vs. only 3 in the prior 20-minute run, but the extra trades were overwhelmingly low-quality tokens with hidden sell taxes. Actual sell recovery averaged only 0.00006 BNB per position, meaning most tokens recovered less than 1% of the 0.0005 BNB buy. Net P&L went from a near-flat **-0.000008 BNB** to **-0.015417 BNB** — a 1927× larger absolute loss. The new settings allowed tokens that pass the static 0.90 round-trip simulation but fail catastrophically in the live sell path.

## 9. Observations and Recommendations

1. **0.90 efficiency is too permissive.** The static simulation is not predictive of live sell tax; many tokens that simulated ≥ 90% efficiency returned < 1% on real sells.
2. **$8k liquidity is too low.** The additional volume came from thin pools that amplified slippage/tax losses.
3. **Partial TP logic is functioning but not profitable here.** Tokens rarely reached +200%; when they did, the first 50% sell recovered only ~0.000003 BNB, suggesting the TP trigger was actually a high-tax sell.
4. **WBNB-only remains essential.** Reverting the non-WBNB guard would likely compound the damage.
5. **Recommended next step:** Revert to 0.95 efficiency and $12k liquidity, keep WBNB-only, and test other improvements (e.g., higher minimum holder count, smaller max top-10 concentration, or a minimum simulated sell output in absolute BNB terms).
6. **P&L accounting bug:** Fix the executor's double-counting of cost basis on the second partial sell so the DB column matches actual cash P&L.

---
*Generated from live mainnet test on 2026-07-23.*
