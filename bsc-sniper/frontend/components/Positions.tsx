"use client";
import { useEffect, useState } from "react";
import { ExternalLink, TrendingUp, TrendingDown } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { api, type Position } from "@/lib/api";
import { bscScanLink, formatPct, shortenAddress, timeAgo } from "@/lib/utils";

interface Props {
  wsEvent: unknown;
}

function calcPnlPct(pos: Position): number {
  const entry = parseFloat(pos.entry_price_bnb);
  const current = parseFloat(pos.current_price_bnb);
  if (!entry || entry === 0) return 0;
  return ((current - entry) / entry) * 100;
}

export function Positions({ wsEvent }: Props) {
  const [positions, setPositions] = useState<Position[]>([]);

  useEffect(() => {
    api.positions().then((r) => setPositions(r.data ?? [])).catch(console.error);
  }, []);

  useEffect(() => {
    if (!wsEvent) return;
    const ev = wsEvent as { type: string; token?: string; price_bnb?: number; pnl_pct?: number; ath_price_bnb?: number };
    if (
      ev.type === "position_opened" ||
      ev.type === "position_opened_simulated" ||
      ev.type === "sell_executed"
    ) {
      api.positions().then((r) => setPositions(r.data ?? [])).catch(console.error);
    } else if (ev.type === "price_update" && ev.token) {
      setPositions((prev) =>
        prev.map((p) =>
          p.token_address === ev.token
            ? {
                ...p,
                current_price_bnb: String(ev.price_bnb ?? p.current_price_bnb),
                ath_price_bnb: String(ev.ath_price_bnb ?? p.ath_price_bnb),
              }
            : p
        )
      );
    }
  }, [wsEvent]);

  const open = positions.filter((p) => p.status === "open" || p.status === "partial");
  const closed = positions.filter((p) => p.status === "closed").slice(0, 10);
  const totalPnl = closed.reduce((sum, p) => sum + (p.realized_pnl_bnb ?? 0), 0);

  return (
    <Card className="h-full">
      <CardHeader className="pb-3">
        <CardTitle className="text-base flex items-center gap-2">
          Positions
          <Badge variant="secondary">{open.length} open</Badge>
          {totalPnl !== 0 && (
            <span className={`ml-auto text-sm font-medium ${totalPnl >= 0 ? "text-green-400" : "text-red-400"}`}>
              Realized P&L: {totalPnl >= 0 ? "+" : ""}{totalPnl.toFixed(6)} BNB
            </span>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <div className="max-h-[400px] overflow-y-auto">
          {positions.length === 0 ? (
            <p className="text-center text-muted-foreground text-sm py-8">No positions</p>
          ) : (
            <table className="w-full text-sm">
              <thead className="sticky top-0 bg-card border-b border-border">
                <tr className="text-muted-foreground text-xs">
                  <th className="text-left px-4 py-2">Token</th>
                  <th className="text-right px-4 py-2">Cost BNB</th>
                  <th className="text-right px-4 py-2">P&L</th>
                  <th className="text-right px-4 py-2">TP1</th>
                  <th className="text-right px-4 py-2">Status</th>
                  <th className="text-right px-4 py-2">Age</th>
                </tr>
              </thead>
              <tbody>
                {[...open, ...closed].map((p) => {
                  const pnlPct = calcPnlPct(p);
                  const isProfit = pnlPct >= 0;
                  return (
                    <tr key={p.id} className="border-b border-border/50 hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-2">
                        <a
                          href={bscScanLink("token", p.token_address)}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1 font-medium text-blue-400 hover:text-blue-300"
                        >
                          {p.token_symbol || shortenAddress(p.token_address)}
                          <ExternalLink className="h-3 w-3" />
                        </a>
                      </td>
                      <td className="px-4 py-2 text-right font-mono text-xs">{p.cost_bnb?.toFixed(6)}</td>
                      <td className={`px-4 py-2 text-right font-mono text-xs font-medium ${isProfit ? "text-green-400" : "text-red-400"}`}>
                        <span className="flex items-center justify-end gap-1">
                          {isProfit ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                          {formatPct(pnlPct)}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-right">
                        {p.tp1_triggered ? (
                          <Badge variant="success">✓</Badge>
                        ) : (
                          <span className="text-muted-foreground text-xs">—</span>
                        )}
                      </td>
                      <td className="px-4 py-2 text-right">
                        <Badge variant={p.status === "open" ? "success" : p.status === "partial" ? "warning" : "secondary"}>
                          {p.status}
                        </Badge>
                      </td>
                      <td className="px-4 py-2 text-right text-xs text-muted-foreground">
                        {timeAgo(p.opened_at)}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
