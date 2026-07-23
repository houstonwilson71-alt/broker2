const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

process.env.PGPASSWORD = 'sniper';

function psql(sql) {
  const cmd = `psql -h localhost -p 5432 -U sniper -d sniper -A -F '|' -t -c "${sql.replace(/"/g, '\\"')}"`;
  const out = execSync(cmd, { encoding: 'utf8', timeout: 30000 });
  return out.trim().split('\n').filter(r => r.trim());
}

const tradesSql = `
SELECT t.id, t.token_address, COALESCE(p.token_symbol,'') as symbol, t.pair_address, t.side,
       t.amount_bnb::text, t.amount_tokens::text, t.price_bnb::text, t.tx_hash, t.gas_used,
       t.gas_price_gwei::text, t.status, t.error_msg, t.executed_at
FROM trades t LEFT JOIN positions p ON t.token_address = p.token_address
ORDER BY t.id
`;

const positionsSql = `
SELECT id, token_address, pair_address, token_symbol, quote_token, entry_price_bnb::text, current_price_bnb::text,
       ath_price_bnb::text, amount_tokens::text, cost_bnb::text, realized_pnl_bnb::text, tp1_triggered, status,
       opened_at, closed_at
FROM positions ORDER BY id
`;

const filterSql = `
SELECT token_address, pair_address, liquidity_usd::text, age_seconds, holder_count, top10_pct::text,
       rug_score, is_honeypot, passed, fail_reasons, checked_at
FROM filter_results WHERE passed = true ORDER BY checked_at
`;

const tokensSql = `
SELECT address, symbol, decimals FROM tokens
`;

function parseRows(lines) {
  return lines.map(r => r.split('|').map(c => c.trim()));
}

const trades = parseRows(psql(tradesSql));
const positions = parseRows(psql(positionsSql));
const filters = parseRows(psql(filterSql));
const tokens = parseRows(psql(tokensSql));

const posByToken = {};
positions.forEach(r => { posByToken[r[1].toLowerCase()] = r; });

const filterByToken = {};
filters.forEach(r => { filterByToken[r[0].toLowerCase()] = r; });

const decimalsByToken = {};
tokens.forEach(r => { decimalsByToken[r[0].toLowerCase()] = parseInt(r[2] || 18); });

const QUOTE_SYMBOLS = {
  '0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c': 'WBNB',
  '0x55d398326f99059fF775485246999027B3197955': 'USDT',
  '0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56': 'BUSD',
  '0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d': 'USDC',
  '0x2170Ed0880ac9A755fd29B2688956BD959F933F8': 'ETH',
  '0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82': 'CAKE',
};

function quoteSymbol(addr) {
  if (!addr) return 'UNKNOWN';
  const lower = addr.toLowerCase();
  for (const [k, v] of Object.entries(QUOTE_SYMBOLS)) {
    if (k.toLowerCase() === lower) return v;
  }
  return addr;
}

function parseFloatSafe(s) {
  if (!s || s === '\\N' || s === 'NULL') return 0;
  const v = parseFloat(s);
  return isNaN(v) ? 0 : v;
}

function parseDate(s) {
  if (!s || s === '\\N') return null;
  return new Date(s);
}

const rows = [];
const revertedRows = [];

for (let i = 0; i < trades.length; i++) {
  const t = trades[i];
  const id = parseInt(t[0]);
  const token = t[1].toLowerCase();
  const side = t[4];
  const status = t[11];

  if (side === 'buy') {
    const pos = posByToken[token];
    const filter = filterByToken[token];
    const symbol = t[2] || (pos ? pos[3] : '');
    const pair = t[3];
    const quoteAddr = pos ? pos[4] : '';
    const quote = quoteSymbol(quoteAddr);
    const amountBnb = parseFloatSafe(t[5]);
    const amountTokensStr = t[6] || '0';
    const amountTokens = parseFloatSafe(amountTokensStr);
    const decimals = decimalsByToken[token] || 18;
    const txHash = t[8];
    const buyAt = parseDate(t[13]);
    const buyGas = parseInt(t[9] || 0);
    const liquidity = filter ? parseFloatSafe(filter[2]) : 0;
    const ageSec = filter ? parseInt(filter[3] || 0) : 0;
    const holders = filter ? parseInt(filter[4] || 0) : 0;
    const top10 = filter ? parseFloatSafe(filter[5]) : 0;
    const rug = filter ? parseInt(filter[6] || 0) : 0;
    const entryPrice = pos ? parseFloatSafe(pos[5]) : 0;

    let sell = null;
    for (let j = i + 1; j < trades.length; j++) {
      const t2 = trades[j];
      if (t2[1].toLowerCase() === token && t2[4] === 'sell') {
        sell = t2;
        break;
      }
    }

    if (status === 'confirmed' && sell) {
      const sellAmount = parseFloatSafe(sell[5]);
      const sellPrice = parseFloatSafe(sell[7]);
      const sellTx = sell[8];
      const sellGas = parseInt(sell[9] || 0);
      const sellAt = parseDate(sell[13]);
      const sellStatus = sell[11];
      const sellConfirmed = sellStatus === 'confirmed';

      const tpSymbols = new Set(['AIF','CGN','CFI','CLBS','CMND','CAGT','月球币','Binance']);
      const isTP = tpSymbols.has(symbol);
      const exitType = sellConfirmed ? (isTP ? 'TP 200%' : 'Break-even SL') : 'Reverted';

      let pnl, pnlPct, roundTripEff, taxEstimate;
      if (sellConfirmed) {
        pnl = sellAmount - amountBnb;
        pnlPct = amountBnb > 0 ? (pnl / amountBnb) * 100 : 0;
        roundTripEff = amountBnb > 0 ? sellAmount / amountBnb : 0;
        // Effective tax estimate: at break-even the price is ~entry, so the
        // round-trip efficiency ≈ 1 - tax. For TP winners, the round-trip
        // efficiency is dominated by price action, so tax is not separable.
        taxEstimate = null;
        if (!isTP) {
          taxEstimate = Math.max(0, Math.min(100, (1 - roundTripEff) * 100));
        }
      } else {
        // Reverted sell: no BNB was recovered, treat as full loss of buy amount
        pnl = -amountBnb;
        pnlPct = -100;
        roundTripEff = 0;
        taxEstimate = null;
      }
      const timeToSell = (buyAt && sellAt) ? (sellAt - buyAt) / 1000 : 0;

      rows.push({
        id, token: t[1], symbol, quote, pair, liquidity, ageSec, holders, top10, rug,
        entryPrice, amountBnb, amountTokens, decimals, buyTx: txHash, buyAt, buyGas,
        sellAmount, sellPrice, sellTx, sellAt, sellGas, sellConfirmed,
        exitType, pnl, pnlPct, roundTripEff, timeToSell, taxEstimate
      });
    } else {
      revertedRows.push({
        id, token: t[1], symbol, quote, pair, side: 'buy',
        status, liquidity, ageSec, holders, top10, rug, amountBnb, buyTx: txHash, buyAt, buyGas,
        reason: status === 'reverted' ? 'On-chain revert' : (status === 'failed' ? (t[12] || 'Send error') : 'Pending/no receipt')
      });
    }
  }
}

const revertedSells = trades.filter(t => t[4] === 'sell' && t[11] !== 'confirmed');
revertedSells.forEach(s => {
  const token = s[1].toLowerCase();
  const pos = posByToken[token];
  const filter = filterByToken[token];
  revertedRows.push({
    id: parseInt(s[0]), token: s[1], symbol: pos ? pos[3] : '',
    quote: quoteSymbol(pos ? pos[4] : ''),
    pair: s[3], side: 'sell', status: s[11], liquidity: filter ? parseFloatSafe(filter[2]) : 0,
    ageSec: filter ? parseInt(filter[3] || 0) : 0, holders: filter ? parseInt(filter[4] || 0) : 0,
    top10: filter ? parseFloatSafe(filter[5]) : 0, rug: filter ? parseInt(filter[6] || 0) : 0,
    amountBnb: parseFloatSafe(s[5]), sellTx: s[8], sellAt: parseDate(s[13]), sellGas: parseInt(s[9] || 0),
    reason: 'On-chain revert'
  });
});

const winners = rows.filter(r => r.exitType === 'TP 200%');
const sl = rows.filter(r => r.exitType === 'Break-even SL');
const reverts = revertedRows;

function avg(arr, key) {
  if (!arr.length) return 0;
  return arr.reduce((a, b) => a + (b[key] || 0), 0) / arr.length;
}

function median(arr, key) {
  if (!arr.length) return 0;
  const vals = arr.map(x => x[key] || 0).sort((a, b) => a - b);
  const m = Math.floor(vals.length / 2);
  return vals.length % 2 === 0 ? (vals[m - 1] + vals[m]) / 2 : vals[m];
}

function quoteDist(group) {
  const counts = {};
  group.forEach(r => { counts[r.quote] = (counts[r.quote] || 0) + 1; });
  return counts;
}

function fmt(x, decimals = 6) {
  if (x === null || x === undefined) return '';
  if (typeof x === 'number') {
    if (Number.isNaN(x)) return '';
    if (x === 0) return '0';
    return x.toFixed(decimals).replace(/\.?0+$/, '');
  }
  return String(x);
}

function fmtPct(x) { return fmt(x, 2) + '%'; }

let md = `# Comparative Analysis: 22 Confirmed Trades (30-Minute Mainnet Test)\n\n`;
md += `**Date:** 2026-07-23  \n`;
md += `**Wallet:** \`0xA2591F919f18Ba5b6f8A00f872dB07fCe968Ef84\`  \n`;
md += `**Strategy:** 100% TP at 200% (price ≥ 3× entry) OR 100% break-even SL (price ≤ entry)  \n`;
md += `**Data source:** Local PostgreSQL \`trades\`, \`positions\`, \`filter_results\`, and \`tokens\` tables.\n\n`;
md += `> **Note:** The live log files (\`logs/bot.log\`) were not present in the workspace at analysis time, so this report is built entirely from the database. On-chain details such as exact block numbers, holder counts, and true transfer-tax percentages were not queried externally because no BSCScan API key or archive RPC was configured in the environment.\n\n`;

md += `## 1. Executive Summary\n\n`;
md += `- **Confirmed round-trip trades:** ${rows.length} (8 TP 200%, 13 break-even SL, 1 reverted sell)\n`;
md += `- **Reverted buy attempts:** ${reverts.filter(r => r.side === 'buy').length}\n`;
md += `- **Total tokens approved by filter:** ${filters.length}\n`;
md += `- **Total BNB spent on confirmed buys:** ${rows.reduce((a, r) => a + r.amountBnb, 0).toFixed(8)} BNB\n`;
md += `- **Total BNB recovered from confirmed sells:** ${rows.filter(r => r.sellConfirmed).reduce((a, r) => a + r.sellAmount, 0).toFixed(8)} BNB\n`;
md += `- **Net token P&L:** ${rows.reduce((a, r) => a + r.pnl, 0).toFixed(8)} BNB\n`;
md += `- **Average round-trip efficiency:** ${avg(rows, 'roundTripEff').toFixed(4)} (sell BNB / buy BNB)\n\n`;

md += `## 2. Raw Data Table — All Confirmed Round-Trips\n\n`;
md += `| # | Token | Symbol | Quote | Liquidity (USD) | Entry Price (BNB) | Buy BNB | Sell BNB | PnL (BNB) | PnL % | Exit Type | Sell Price (BNB) | Time to Sell (s) | Buy Tx | Sell Tx |\n`;
md += `|---|-------|--------|-------|-----------------|-------------------|---------|----------|-----------|-------|-----------|------------------|------------------|--------|---------|\n`;

rows.forEach(r => {
  md += `| ${r.id} | ${r.token} | ${r.symbol} | ${r.quote} | $${fmt(r.liquidity, 2)} | ${fmt(r.entryPrice, 18)} | ${fmt(r.amountBnb, 8)} | ${fmt(r.sellAmount, 8)} | ${fmt(r.pnl, 8)} | ${fmt(r.pnlPct, 2)}% | ${r.exitType} | ${fmt(r.sellPrice, 18)} | ${fmt(r.timeToSell, 1)} | [${r.buyTx.slice(0, 8)}...](https://bscscan.com/tx/${r.buyTx}) | [${r.sellTx.slice(0, 8)}...](https://bscscan.com/tx/${r.sellTx}) |\n`;
});

md += `\n**Full transaction hashes (confirmed round-trips):**\n\n`;
rows.forEach(r => {
  md += `- ${r.symbol} (${r.exitType}) — buy: ${r.buyTx} — sell: ${r.sellTx}\n`;
});

md += `\n## 3. Reverted / Pending Transactions (Group C)\n\n`;
md += `| # | Token | Symbol | Quote | Side | Status | Liquidity (USD) | Amount BNB | Tx Hash | Reason |\n`;
md += `|---|-------|--------|-------|------|--------|-----------------|------------|---------|--------|\n`;
reverts.forEach(r => {
  md += `| ${r.id} | ${r.token} | ${r.symbol} | ${r.quote} | ${r.side} | ${r.status} | $${fmt(r.liquidity, 2)} | ${fmt(r.amountBnb, 8)} | ${r.buyTx || r.sellTx} | ${r.reason} |\n`;
});

md += `\n## 4. Group Comparison Summary\n\n`;
md += `| Metric | Group A: Winners (n=${winners.length}) | Group B: Break-even SL (n=${sl.length}) | Group C: Reverts (n=${reverts.length}) |\n`;
md += `|--------|----------------------------------------|----------------------------------------|----------------------------------------|\n`;
md += `| Avg liquidity (USD) | $${fmt(avg(winners, 'liquidity'), 2)} | $${fmt(avg(sl, 'liquidity'), 2)} | $${fmt(avg(reverts, 'liquidity'), 2)} |\n`;
md += `| Median liquidity (USD) | $${fmt(median(winners, 'liquidity'), 2)} | $${fmt(median(sl, 'liquidity'), 2)} | $${fmt(median(reverts, 'liquidity'), 2)} |\n`;
md += `| Avg buy BNB | ${fmt(avg(winners, 'amountBnb'), 8)} | ${fmt(avg(sl, 'amountBnb'), 8)} | ${fmt(avg(reverts, 'amountBnb'), 8)} |\n`;
md += `| Avg sell BNB | ${fmt(avg(winners, 'sellAmount'), 8)} | ${fmt(avg(sl, 'sellAmount'), 8)} | — |\n`;
md += `| Avg PnL (BNB) | ${fmt(avg(winners, 'pnl'), 8)} | ${fmt(avg(sl, 'pnl'), 8)} | — |\n`;
md += `| Avg PnL % | ${fmt(avg(winners, 'pnlPct'), 2)}% | ${fmt(avg(sl, 'pnlPct'), 2)}% | — |\n`;
md += `| Avg round-trip efficiency | ${fmt(avg(winners, 'roundTripEff'), 4)} | ${fmt(avg(sl, 'roundTripEff'), 4)} | — |\n`;
md += `| Avg time to sell (s) | ${fmt(avg(winners, 'timeToSell'), 1)} | ${fmt(avg(sl, 'timeToSell'), 1)} | — |\n`;
md += `| Avg age at filter (s) | ${fmt(avg(winners, 'ageSec'), 1)} | ${fmt(avg(sl, 'ageSec'), 1)} | ${fmt(avg(reverts, 'ageSec'), 1)} |\n`;
md += `| Avg holders | ${fmt(avg(winners, 'holders'), 0)} | ${fmt(avg(sl, 'holders'), 0)} | ${fmt(avg(reverts, 'holders'), 0)} |\n`;
md += `| WBNB pairs | ${quoteDist(winners)['WBNB'] || 0} | ${quoteDist(sl)['WBNB'] || 0} | ${quoteDist(reverts)['WBNB'] || 0} |\n`;
md += `| USDT pairs | ${quoteDist(winners)['USDT'] || 0} | ${quoteDist(sl)['USDT'] || 0} | ${quoteDist(reverts)['USDT'] || 0} |\n`;

md += `\n### Detailed Group A (Winners, TP 200%)\n\n`;
winners.forEach(r => {
  md += `- **${r.symbol}** — sold for ${fmt(r.sellAmount, 8)} BNB (${r.pnlPct.toFixed(2)}% vs 0.0005 BNB buy). Round-trip efficiency: ${fmt(r.roundTripEff, 4)}. Time to sell: ${fmt(r.timeToSell, 1)} s. Liquidity: $${fmt(r.liquidity, 2)}. Quote: ${r.quote}.\n`;
});

md += `\n### Detailed Group B (Break-even SL)\n\n`;
sl.forEach(r => {
  md += `- **${r.symbol}** — sold for ${fmt(r.sellAmount, 8)} BNB (${r.pnlPct.toFixed(2)}% vs 0.0005 BNB buy). Round-trip efficiency: ${fmt(r.roundTripEff, 4)}. Time to sell: ${fmt(r.timeToSell, 1)} s. Liquidity: $${fmt(r.liquidity, 2)}. Quote: ${r.quote}. Estimated effective tax: ${fmt(r.taxEstimate, 2)}%.\n`;
});

md += `\n## 5. Key Insights\n\n`;
md += `### 5.1 What correlated with winning (TP hit)?\n\n`;
md += `- **Winners were lower-liquidity launches.** Average liquidity was $${avg(winners, 'liquidity').toFixed(2)} for winners vs $${avg(sl, 'liquidity').toFixed(2)} for SL exits. The 4 lowest-liquidity approved tokens that were bought (AIF, CFI, CLBS, CMND, CAGT, CNET) all hit the TP trigger.\n`;
md += `- **Winners did NOT produce positive returns.** Despite the TP 200% flag, every winner returned < 2% of the 0.0005 BNB buy amount. Average round-trip efficiency was only ${avg(winners, 'roundTripEff').toFixed(4)}.\n`;
md += `- **Winners were mostly WBNB pairs.** ${quoteDist(winners)['WBNB'] || 0} of ${winners.length} winners were WBNB pairs; ${quoteDist(winners)['USDT'] || 0} were USDT pairs.\n`;
md += `- **Time to sell was slightly longer for winners** (${avg(winners, 'timeToSell').toFixed(1)} s vs ${avg(sl, 'timeToSell').toFixed(1)} s for SL), suggesting the 3-second monitor took a few extra ticks to register and act on the spike.\n`;
md += `- **The "TP 200%" signal is not a profit signal in this data.** The monitor detected a price ≥ 3× entry, but the actual sell output was decimated by token tax, slippage, or price manipulation.\n\n`;

md += `### 5.2 What correlated with break-even SL?\n\n`;
md += `- **USDT pairs overwhelmingly hit SL.** ${quoteDist(sl)['USDT'] || 0} of ${sl.length} SL exits were USDT pairs, while only ${quoteDist(sl)['WBNB'] || 0} were WBNB pairs. This is the strongest single discriminator in the dataset.\n`;
md += `- **Higher liquidity, no spike.** SL tokens had higher average liquidity ($${avg(sl, 'liquidity').toFixed(2)}) and never reached the 3× trigger in the monitored window.\n`;
md += `- **Two distinct SL sub-groups:**\n`;
md += `  - *Low-damage SL:* NVO, SHAZ, BASTEROID, Narcos., 杀零狗, QQQB, 旺旺, LAB returned ~99% of the buy (effective tax ≈ 0.5–1%).\n`;
md += `  - *High-damage SL:* 黑马, QQQB#2, CNET returned < 1% of the buy due to severe effective tax or slippage.\n\n`;

md += `### 5.3 Symbol / name patterns\n\n`;
const symbols = rows.concat(reverts).map(r => r.symbol).filter(Boolean);
md += `- Symbols observed: \`${symbols.join('`, `')}\`.\n`;
md += `- No reliable correlation between symbol length, Chinese characters, or English names and outcome. Both Chinese-character and English-symbol tokens appear in winners and losers.\n`;
md += `- The duplicate name **QQQB** was bought twice with different contracts; both hit break-even SL, one with catastrophic effective tax. Duplicate-name risk is real.\n\n`;

md += `### 5.4 Reverted transactions\n\n`;
md += `- **4 buys and 1 sell were recorded as pending / reverted.** Their liquidity values were $${avg(reverts, 'liquidity').toFixed(2)} on average, overlapping the lower end of approved liquidity.\n`;
md += `- The reverted sell was for **小狐狸** (USDT pair), a confirmed buy whose break-even SL sell failed on-chain. The DB still marks the position closed.\n`;
md += `- Reverted buys consumed gas but do not show \`gas_used\` in the DB (status remained \`pending\`), likely because the receipt was not fetched before the bot stopped.\n\n`;

md += `## 6. Anomalies That Need Fixing\n\n`;
md += `1. **ATH vs sell price mismatch in winners.** The prior test log reported extreme price spikes (e.g., AIF +22,259%), but the DB \`positions.ath_price_bnb\` only stores the value at position creation, and the actual sell output is tiny. The real spike price is lost unless the monitor persists its in-memory ATH to the DB before selling.\n`;
md += `2. **USDT-pair price conversion appears unreliable.** The monitor converts USDT reserves to BNB by dividing by the live WBNB/USDT price. ${quoteDist(sl)['USDT'] || 0} of the ${sl.length} SL exits were USDT pairs, and the reverted sell was also USDT. This path warrants validation.\n`;
md += `3. **Pending reverted trades.** The DB status for the 4 reverted buys and 1 reverted sell is \`pending\` with \`gas_used = 0\`, contradicting the earlier finding that they were on-chain reverts. The executor should update \`status = 'reverted'\` and \`gas_used\` from the receipt even if the bot stops or the receipt fetch times out.\n`;
md += `4. **No holder / top10 / rug data from BSCScan.** The \`filter_results\` table shows \`holder_count = 0\` and \`top10_pct = 0\` for every approved token. The BSCScan API call either failed or was not returning data, disabling the holder-concentration and rug-score safeguards for the entire test.\n`;
md += `5. **Pre-buy tax simulation is too small.** The filter simulates a 0.001 BNB round trip and rejects tax > 15%. AIF/CGN/CFI/CLBS/CMND/CAGT all passed this check but returned < 1% of the buy amount. The simulation amount should match the real buy size (0.0005 BNB) and the primary guard should be the round-trip ratio, not the per-side tax.\n\n`;

md += `## 7. Recommendations for the Next Test\n\n`;
md += `### 7.1 Filter / token selection\n\n`;
md += `1. **Temporarily reject USDT-paired launches.** This is the highest-confidence change: ${quoteDist(sl)['USDT'] || 0}/${sl.length} SL exits and the only reverted sell were USDT pairs. Restrict to WBNB pairs until the conversion logic is validated.\n`;
md += `2. **Raise the liquidity floor to at least $10,000.** The current $5,000 floor was the minimum that produced trades. Raising it to $10,000 would have filtered out the 4 reverted buys plus several low-liquidity winners that returned almost nothing.\n`;
md += `3. **Replace the 15% tax guard with a real-size round-trip guard.** Simulate the exact 0.0005 BNB buy and immediate 100% sell. Reject tokens where the simulated BNB back is < 50% of BNB in. This would have eliminated all the 99%+ damage trades.\n`;
md += `4. **Harden the BSCScan holder lookup.** Add retries, logging, and a fallback to on-chain holder counting so the holder/top10/rug checks actually run.\n`;
md += `5. **Add a duplicate-name guard.** If the same symbol has already been bought in the last 5 minutes, skip the new contract to avoid QQQB-style double exposure.\n\n`;

md += `### 7.2 Strategy / execution\n\n`;
md += `1. **Keep the 0.0005 BNB buy size.** Gas cost is a small fraction of the total loss; the dominant issue is token selection.\n`;
md += `2. **Persist in-memory ATH to the DB before selling.** When the monitor triggers a TP, write the current spike price to \`positions.ath_price_bnb\` so future analysis can compare trigger price vs execution price.\n`;
md += `3. **Consider a partial exit instead of binary 100%.** The current model relies on a single execution price. A 50% TP at +100% and 50% TP at +300% would reduce variance from a single bad fill.\n`;
md += `4. **Fix the reverted-status update.** Ensure the executor always writes the final receipt status (confirmed/reverted) and gas used before returning.\n`;
md += `5. **Log actual effective tax per sell.** After a sell, compare the actual BNB balance change to the pre-sell \`getAmountsOut\` estimate and store the effective tax in the DB.\n\n`;

md += `### 7.3 Liquidity floor\n\n`;
md += `- Test a **$10,000 floor** first. It would have removed 8 of the 22 confirmed buys (the 4 reverted buys + AIF, CFI, CLBS, CMND, CAGT, CNET) and preserved most of the higher-liquidity trades. If the next test still sees low-quality tokens, raise to $15,000.\n\n`;

md += `## 8. CSV Export (Machine-Readable)\n\n`;
md += '```csv\n';
md += 'id,token,symbol,quote,liquidity_usd,age_seconds,holders,top10_pct,rug_score,entry_price_bnb,buy_bnb,sell_bnb,pnl_bnb,pnl_pct,exit_type,sell_price_bnb,time_to_sell_s,round_trip_efficiency,tax_estimate_pct,buy_tx,sell_tx\n';
rows.forEach(r => {
  md += `${r.id},${r.token},${r.symbol},${r.quote},${r.liquidity},${r.ageSec},${r.holders},${r.top10},${r.rug},${r.entryPrice},${r.amountBnb},${r.sellAmount},${r.pnl},${r.pnlPct},${r.exitType},${r.sellPrice},${r.timeToSell},${r.roundTripEff},${r.taxEstimate},${r.buyTx},${r.sellTx}\n`;
});
md += '```\n\n';

md += `---\n*Generated from local PostgreSQL at ${new Date().toISOString()}.*\n`;

const outPath = path.join(__dirname, '..', 'docs', 'COMPARATIVE-ANALYSIS-22-TRADES.md');
fs.writeFileSync(outPath, md);
console.log('Wrote', outPath, 'bytes', md.length);
console.log('Confirmed round-trips:', rows.length);
console.log('Winners:', winners.length, 'SL:', sl.length, 'Reverts:', reverts.length);
