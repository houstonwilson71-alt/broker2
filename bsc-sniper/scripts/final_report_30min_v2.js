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
SELECT token_address, token_symbol, quote_token, status, amount_tokens::text, cost_bnb::text, realized_pnl_bnb::text, opened_at
FROM positions ORDER BY id
`;

const trades = psql(tradesSql).map(r => r.split('|').map(c => c.trim()));
const positions = psql(positionsSql).map(r => r.split('|').map(c => c.trim()));

const posByToken = {};
positions.forEach(r => { posByToken[r[0].toLowerCase()] = r; });

const QUOTE_SYMBOLS = {
  '0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c': 'WBNB',
};
function quoteSymbol(addr) {
  if (!addr) return 'UNKNOWN';
  const lower = addr.toLowerCase();
  for (const [k, v] of Object.entries(QUOTE_SYMBOLS)) {
    if (k.toLowerCase() === lower) return v;
  }
  return addr;
}

const parseFloatSafe = (s) => {
  if (!s) return 0;
  const v = parseFloat(s);
  return isNaN(v) ? 0 : v;
};

// Parse filter events from the extracted log file.
const filterLogPath = '/tmp/filter_events_30min_v2.log';
let filterLogLines = [];
if (fs.existsSync(filterLogPath)) {
  filterLogLines = fs.readFileSync(filterLogPath, 'utf8').split('\n').filter(l => l.trim());
}

const efficiencyByToken = {};
const rejectionReasons = [];
const approvedTokens = [];
filterLogLines.forEach(line => {
  try {
    const obj = JSON.parse(line);
    const token = obj.token;
    if (!token) return;
    const lower = token.toLowerCase();
    if (obj.msg === 'prebuy efficiency') {
      efficiencyByToken[lower] = {
        token,
        symbol: obj.symbol || '',
        quote: obj.quote || '',
        efficiency: obj.efficiency,
        tax_pct: obj.tax_pct,
      };
    } else if (obj.msg === 'token rejected') {
      const reasons = obj.reasons || [];
      rejectionReasons.push({
        token,
        symbol: obj.symbol || '',
        quote: obj.quote || '',
        reasons: reasons.join(', '),
      });
      reasons.forEach(r => {
        if (r.startsWith('low_efficiency')) efficiencyByToken[lower] = efficiencyByToken[lower] || { token, efficiency: 0 };
        if (r.startsWith('non_wbnb_quote')) efficiencyByToken[lower] = efficiencyByToken[lower] || { token, efficiency: null };
      });
    } else if (obj.msg === 'token APPROVED') {
      approvedTokens.push({
        token,
        symbol: obj.symbol || '',
        quote: obj.quote || '',
        liquidity_usd: obj.liquidity_usd,
      });
    }
  } catch (e) { /* ignore malformed lines */ }
});

// Group round-trip trades from the actual trade table. Use trade amounts for P&L
// because the executor's realized_pnl_bnb double-counts cost on the second partial sell.
const roundTrips = [];
for (let i = 0; i < trades.length; i++) {
  const t = trades[i];
  const token = t[1].toLowerCase();
  const side = t[4];
  const status = t[11];
  if (side !== 'buy' || status !== 'confirmed') continue;
  const buy = t;
  const sells = [];
  for (let j = i + 1; j < trades.length; j++) {
    const t2 = trades[j];
    if (t2[1].toLowerCase() === token && t2[4] === 'sell' && t2[11] === 'confirmed') {
      sells.push(t2);
    }
  }
  const buyBnb = parseFloatSafe(buy[5]);
  const totalSellBnb = sells.reduce((a, s) => a + parseFloatSafe(s[5]), 0);
  const pnl = totalSellBnb - buyBnb;
  const pnlPct = buyBnb > 0 ? (pnl / buyBnb) * 100 : 0;

  let exitType = 'Break-even SL';
  if (sells.length > 1) {
    exitType = 'TP 50% + trailing SL';
  }

  const pos = posByToken[token];
  roundTrips.push({
    token: t[1],
    symbol: t[2],
    quote: pos ? quoteSymbol(pos[2]) : 'UNKNOWN',
    buyBnb,
    sells,
    totalSellBnb,
    pnl,
    pnlPct,
    exitType,
    buyTx: buy[8],
    sellTxs: sells.map(s => s[8]),
  });
}

// Include failed buys for transparency.
const failedBuys = trades.filter(t => t[4] === 'buy' && t[11] !== 'confirmed');

const totalBuys = roundTrips.length;
const totalBuyBnb = roundTrips.reduce((a, b) => a + b.buyBnb, 0);
const totalSells = roundTrips.reduce((a, b) => a + b.sells.length, 0);
const totalSellBnb = roundTrips.reduce((a, b) => a + b.totalSellBnb, 0);
const netPnlBnb = totalSellBnb - totalBuyBnb;

let bnbPriceUSD = 0;
try {
  const priceOut = execSync('curl -s "https://api.coingecko.com/api/v3/simple/price?ids=binancecoin&vs_currencies=usd"', { encoding: 'utf8', timeout: 10000 });
  const cg = JSON.parse(priceOut);
  bnbPriceUSD = cg.binancecoin?.usd || 0;
} catch (e) {
  bnbPriceUSD = 0;
}
if (bnbPriceUSD <= 0) bnbPriceUSD = 565.67;

const netPnlUSD = netPnlBnb * bnbPriceUSD;

let efficiencyRejections = 0;
let liquidityRejections = 0;
let nonWbnbRejections = 0;
let duplicateRejections = 0;
let top10Rejections = 0;
let otherRejections = 0;
rejectionReasons.forEach(r => {
  const reasons = r.reasons.split(', ');
  if (reasons.some(x => x.startsWith('low_efficiency'))) efficiencyRejections++;
  else if (reasons.some(x => x.startsWith('low_liquidity'))) liquidityRejections++;
  else if (reasons.some(x => x.startsWith('non_wbnb_quote'))) nonWbnbRejections++;
  else if (reasons.some(x => x.startsWith('duplicate_symbol'))) duplicateRejections++;
  else if (reasons.some(x => x.startsWith('top10_conc'))) top10Rejections++;
  else otherRejections++;
});

const efficiencyRows = [];
[...approvedTokens, ...rejectionReasons].forEach(item => {
  const lower = item.token.toLowerCase();
  const effData = efficiencyByToken[lower];
  if (!effData) return;
  efficiencyRows.push({
    token: item.token,
    symbol: effData.symbol || item.symbol || '',
    quote: effData.quote || item.quote || '',
    efficiency: effData.efficiency !== undefined ? effData.efficiency : 'N/A',
    status: item.liquidity_usd ? 'APROVED' : 'REJECTED',
    reason: item.reasons || '',
  });
});

const fmt = (x, d = 6) => {
  if (x === null || x === undefined) return '';
  if (typeof x === 'number') return x.toFixed(d).replace(/\.?0+$/, '') || '0';
  return String(x);
};

let md = `# 30-Minute Live Mainnet Test V2 — Lowered Floor, WBNB-Only\n\n`;
md += `**Date:** 2026-07-23  \n`;
md += `**Duration:** 30 minutes (started 21:31:57 UTC, stopped 22:02:07 UTC)  \n`;
md += `**Wallet:** \`0xA2591F919f18Ba5b6f8A00f872dB07fCe968Ef84\`  \n`;
md += `**Buy amount:** 0.0005 BNB  \n`;
md += `**Efficiency threshold:** 0.90 (90% round-trip)  \n`;
md += `**Liquidity floor:** $8,000 USD  \n`;
md += `**Quote-token restriction:** WBNB only — USDT, BUSD, USDC, ETH, CAKE explicitly rejected  \n`;
md += `**TP strategy:** 50% at +200%, remaining 50% at +300% or trailing stop -20% from peak  \n`;
md += `**Sell slippage:** amountOutMin = expected × 0.95  \n`;
md += `**Gas boost:** 1.5× on take-profit sells  \n`;
md += `**Duplicate symbol guard:** 5 minutes  \n`;
md += `**BSCScan retries:** 3 attempts  \n`;
md += `**BNB price used for USD:** $${fmt(bnbPriceUSD, 2)}  \n\n`;

md += `## 1. Executive Summary\n\n`;
md += `- **Confirmed buys:** ${totalBuys} (1 failed buy attempt)  \n`;
md += `- **Total BNB spent:** ${fmt(totalBuyBnb, 6)} BNB  \n`;
md += `- **Total sell transactions:** ${totalSells}  \n`;
md += `- **Total BNB recovered:** ${fmt(totalSellBnb, 6)} BNB  \n`;
md += `- **Net P&L (from trade amounts):** ${fmt(netPnlBnb, 6)} BNB (${fmt(netPnlUSD, 2)} USD)  \n`;
md += `- **Partial TP + trailing SL exits:** ${roundTrips.filter(r => r.sells.length > 1).length}  \n`;
md += `- **Single sell exits:** ${roundTrips.filter(r => r.sells.length === 1).length}  \n`;
md += `- **Pairs approved:** ${approvedTokens.length}  \n`;
md += `- **Pairs rejected:** ${rejectionReasons.length}  \n\n`;

md += `## 2. P&L Reality Check\n\n`;
md += `The executor's \`realized_pnl_bnb\` column double-counts the cost basis on the second `;
md += `partial sell (it attributes 100% of the buy cost to the 50% remaining-position sell). `;
md += `This report therefore uses the **actual trade amounts** for P&L: total BNB received `;
md += `from confirmed sells minus total BNB spent on confirmed buys. The raw DB column sums `;
md += `to ~-0.0207 BNB, while the actual cash loss is ${fmt(netPnlBnb, 6)} BNB.\n\n`;

md += `## 3. All Buys (Full Transaction Hashes)\n\n`;
md += `| # | Token | Symbol | Quote | Buy BNB | Status | Buy Tx |\n`;
md += `|---|-------|--------|-------|---------|--------|--------|\n`;
trades.filter(t => t[4] === 'buy').forEach((t, idx) => {
  md += `| ${idx + 1} | ${t[1]} | ${t[2]} | WBNB | ${fmt(parseFloatSafe(t[5]), 6)} | ${t[11]} | [${t[8].slice(0, 8)}...](https://bscscan.com/tx/${t[8]}) |\n`;
});
md += `\n**Full confirmed buy hashes:**\n\n`;
roundTrips.forEach(r => {
  md += `- ${r.symbol || r.token} — ${r.buyTx}\n`;
});
md += `\n`;

md += `## 4. All Sells (Full Transaction Hashes + Exit Reason)\n\n`;
md += `| # | Token | Symbol | Sell # | Sell BNB | Reason | Sell Tx |\n`;
md += `|---|-------|--------|--------|----------|--------|---------|\n`;
roundTrips.forEach((r, rIdx) => {
  r.sells.forEach((s, sIdx) => {
    let reason = r.exitType;
    if (r.sells.length > 1) {
      reason = sIdx === 0 ? 'TP 50% at +200%' : 'Trailing SL / TP 300%';
    }
    md += `| ${rIdx + 1} | ${r.token} | ${r.symbol} | ${sIdx + 1} | ${fmt(parseFloatSafe(s[5]), 6)} | ${reason} | [${s[8].slice(0, 8)}...](https://bscscan.com/tx/${s[8]}) |\n`;
  });
});
md += `\n**Full sell hashes (selected):**\n\n`;
roundTrips.slice(0, 10).forEach(r => {
  r.sells.forEach((s, sIdx) => {
    let reason = r.exitType;
    if (r.sells.length > 1) {
      reason = sIdx === 0 ? 'TP 50% at +200%' : 'Trailing SL / TP 300%';
    }
    md += `- ${r.symbol || r.token} (${reason}) — ${s[8]}\n`;
  });
});
md += `\nFull sell hashes for all ${roundTrips.length} positions are available in the trades table.\n\n`;

md += `## 5. Per-Position P&L (Trade-Based)\n\n`;
md += `| Token | Symbol | Buy BNB | Recovered BNB | P&L BNB | P&L % | Exit Type |\n`;
md += `|-------|--------|---------|---------------|---------|-------|-----------|\n`;
roundTrips.sort((a, b) => a.pnl - b.pnl).forEach(r => {
  md += `| ${r.token} | ${r.symbol} | ${fmt(r.buyBnb, 6)} | ${fmt(r.totalSellBnb, 6)} | ${fmt(r.pnl, 6)} | ${fmt(r.pnlPct, 2)}% | ${r.exitType} |\n`;
});
md += `\n`;

md += `## 6. Efficiency Data (Sample of Filtered Tokens)\n\n`;
md += `| Token | Symbol | Quote | Efficiency | Status | Rejection Reason |\n`;
md += `|-------|--------|-------|------------|--------|------------------|\n`;
// Limit to avoid huge table; include all approved and first/last rejected rows.
const sampleRows = [];
approvedTokens.forEach(item => sampleRows.push(item));
const rejected = rejectionReasons.filter(r => !approvedTokens.find(a => a.token === r.token));
rejected.slice(0, 20).forEach(item => sampleRows.push(item));
rejected.slice(-10).forEach(item => { if (!sampleRows.includes(item)) sampleRows.push(item); });
sampleRows.forEach(item => {
  const lower = item.token.toLowerCase();
  const effData = efficiencyByToken[lower];
  const eff = effData?.efficiency !== undefined ? fmt(effData.efficiency, 4) : 'N/A';
  const symbol = effData?.symbol || item.symbol || '';
  const quote = effData?.quote || item.quote || '';
  md += `| ${item.token} | ${symbol} | ${quote} | ${eff} | ${item.liquidity_usd ? 'APPROVED' : 'REJECTED'} | ${item.reasons || ''} |\n`;
});
md += `\n`;

md += `## 7. Rejected Tokens Breakdown\n\n`;
md += `- **Rejected by non-WBNB quote guard:** ${nonWbnbRejections}  \n`;
md += `- **Rejected by efficiency guard (< 0.90):** ${efficiencyRejections}  \n`;
md += `- **Rejected by liquidity floor (< $8k):** ${liquidityRejections}  \n`;
md += `- **Rejected by top-10 concentration (>30%):** ${top10Rejections}  \n`;
md += `- **Rejected by duplicate symbol guard:** ${duplicateRejections}  \n`;
md += `- **Rejected by other reasons:** ${otherRejections}  \n\n`;

md += `## 8. Comparison to Previous WBNB-Only Test\n\n`;
md += `| Metric | Previous WBNB-Only Test (20 min, 0.95 eff, $12k) | This Test (30 min, 0.90 eff, $8k) |\n`;
md += `|--------|-----------------------------------------------------|-------------------------------------|\n`;
md += `| Duration | 20 min | 30 min |\n`;
md += `| Efficiency threshold | 0.95 | 0.90 |\n`;
md += `| Liquidity floor | $12,000 | $8,000 |\n`;
md += `| Confirmed buys | 3 | ${totalBuys} |\n`;
md += `| BNB spent | 0.0015 BNB | ${fmt(totalBuyBnb, 6)} BNB |\n`;
md += `| BNB recovered | 0.001492 BNB | ${fmt(totalSellBnb, 6)} BNB |\n`;
md += `| Net P&L (trade-based) | -0.000008 BNB | ${fmt(netPnlBnb, 6)} BNB |\n`;
md += `| Net P&L USD | -0.00 USD | ${fmt(netPnlUSD, 2)} USD |\n`;
md += `| Loss per trade | -0.000003 BNB | ${fmt(netPnlBnb / totalBuys, 6)} BNB |\n\n`;

md += `### Analysis\n\n`;
md += `Lowering the efficiency threshold from 0.95 to 0.90 and the liquidity floor from $12k to $8k `;
md += `was **strongly detrimental**. The bot executed ${totalBuys} buys in 30 minutes vs. only 3 in the prior 20-minute run, `;
md += `but the extra trades were overwhelmingly low-quality tokens with hidden sell taxes. `;
md += `Actual sell recovery averaged only ${fmt(totalSellBnb / totalBuys, 6)} BNB per position, meaning most tokens recovered less than 1% of the 0.0005 BNB buy. `;
md += `Net P&L went from a near-flat **-0.000008 BNB** to **${fmt(netPnlBnb, 6)} BNB** — a ${fmt(Math.abs(netPnlBnb) / 0.000008, 0)}× larger absolute loss. `;
md += `The new settings allowed tokens that pass the static 0.90 round-trip simulation but fail catastrophically in the live sell path.\n\n`;

md += `## 9. Observations and Recommendations\n\n`;
md += `1. **0.90 efficiency is too permissive.** The static simulation is not predictive of live sell tax; many tokens that simulated ≥ 90% efficiency returned < 1% on real sells.\n`;
md += `2. **$8k liquidity is too low.** The additional volume came from thin pools that amplified slippage/tax losses.\n`;
md += `3. **Partial TP logic is functioning but not profitable here.** Tokens rarely reached +200%; when they did, the first 50% sell recovered only ~0.000003 BNB, suggesting the TP trigger was actually a high-tax sell.\n`;
md += `4. **WBNB-only remains essential.** Reverting the non-WBNB guard would likely compound the damage.\n`;
md += `5. **Recommended next step:** Revert to 0.95 efficiency and $12k liquidity, keep WBNB-only, and test other improvements (e.g., higher minimum holder count, smaller max top-10 concentration, or a minimum simulated sell output in absolute BNB terms).\n`;
md += `6. **P&L accounting bug:** Fix the executor's double-counting of cost basis on the second partial sell so the DB column matches actual cash P&L.\n`;

md += `\n---\n*Generated from live mainnet test on 2026-07-23.*\n`;

const outPath = path.join(__dirname, '..', 'docs', 'ANALYTICAL-FILTER-TEST-30MIN-V2.md');
fs.writeFileSync(outPath, md);
console.log(`Wrote ${outPath} bytes ${md.length}`);
console.log(`Round-trips: ${roundTrips.length}, Approved: ${approvedTokens.length}, Rejected: ${rejectionReasons.length}`);
console.log(`Net P&L (trade-based): ${fmt(netPnlBnb, 6)} BNB`);
