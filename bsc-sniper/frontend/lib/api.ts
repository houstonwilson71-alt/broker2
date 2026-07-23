const BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface BotStatus {
  running: boolean;
  started_at: string | null;
  stopped_at: string | null;
  pairs_seen: number;
  pairs_passed: number;
  trades_total: number;
  ws_clients: number;
}

export interface Token {
  id: number;
  address: string;
  symbol: string;
  name: string;
  decimals: number;
  pair_address: string;
  block_number: number;
  created_at: string;
}

export interface Trade {
  id: number;
  token_address: string;
  pair_address: string;
  side: "buy" | "sell";
  amount_bnb: number;
  amount_tokens: string;
  price_bnb: string;
  tx_hash: string;
  gas_used: number;
  gas_price_gwei: number;
  status: "pending" | "confirmed" | "failed";
  error_msg: string;
  executed_at: string;
}

export interface Position {
  id: number;
  token_address: string;
  pair_address: string;
  token_symbol: string;
  entry_price_bnb: string;
  current_price_bnb: string;
  ath_price_bnb: string;
  amount_tokens: string;
  cost_bnb: number;
  realized_pnl_bnb: number;
  tp1_triggered: boolean;
  status: "open" | "closed" | "partial";
  opened_at: string;
  closed_at: string | null;
}

export interface BotConfig {
  live_trading_enabled: boolean;
  buy_amount_bnb: number;
  slippage_bps: number;
  min_liquidity_usd: number;
  max_age_sec: number;
  min_holders: number;
  max_top10_pct: number;
  max_rug_score: number;
  take_profit_1_pct: number;
  trailing_stop_pct: number;
}

async function req<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...opts,
    headers: { "Content-Type": "application/json", ...opts?.headers },
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API ${path}: ${res.status} ${text}`);
  }
  return res.json();
}

export const api = {
  health: () => req<{ status: string; bot_running: boolean }>("/api/health"),
  botStatus: () => req<BotStatus>("/api/bot/status"),
  botStart: () => req<{ status: string }>("/api/bot/start", { method: "POST" }),
  botStop: () => req<{ status: string }>("/api/bot/stop", { method: "POST" }),
  tokens: (limit = 50) => req<{ data: Token[]; count: number }>(`/api/tokens?limit=${limit}`),
  trades: (limit = 50) => req<{ data: Trade[]; count: number }>(`/api/trades?limit=${limit}`),
  positions: (status?: string) =>
    req<{ data: Position[]; count: number }>(`/api/positions${status ? `?status=${status}` : ""}`),
  getConfig: () => req<BotConfig>("/api/config"),
  updateConfig: (cfg: Partial<BotConfig>) =>
    req<BotConfig>("/api/config", { method: "PUT", body: JSON.stringify(cfg) }),
};

export function createWebSocket(onMessage: (data: unknown) => void): WebSocket {
  const wsBase = BASE.replace(/^http/, "ws");
  const ws = new WebSocket(`${wsBase}/api/ws`);
  ws.onmessage = (e) => {
    try {
      onMessage(JSON.parse(e.data));
    } catch {
      // ignore parse errors
    }
  };
  return ws;
}
