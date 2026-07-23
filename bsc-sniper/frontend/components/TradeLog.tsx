"use client";
import { useEffect, useState } from "react";
import { ExternalLink } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { api, type Trade } from "@/lib/api";
import { bscScanLink, shortenAddress, shortenHash, timeAgo } from "@/lib/utils";

interface Props {
  wsEvent: unknown;
}

export function TradeLog({ wsEvent }: Props) {
  const [trades, setTrades] = useState<Trade[]>([]);

  useEffect(() => {
    api.trades(50).then((r) => setTrades(r.data ?? [])).catch(console.error);
  }, []);

  useEffect(() => {
    if (!wsEvent) return;
    const ev = wsEvent as { type: string; tx_hash?: string; token?: string; symbol?: string; bnb_received?: number; pnl_bnb?: number; pct?: number };
    if (ev.type === "sell_executed" || ev.type === "position_opened") {
      api.trades(50).then((r) => setTrades(r.data ?? [])).catch(console.error);
    }
  }, [wsEvent]);

  function statusVariant(s: string): "success" | "warning" | "danger" | "secondary" {
    if (s === "confirmed") return "success";
    if (s === "pending") return "warning";
    if (s === "failed") return "danger";
    return "secondary";
  }

  return (
    <Card className="h-full">
      <CardHeader className="pb-3">
        <CardTitle className="text-base flex items-center gap-2">
          Trade Log
          <Badge variant="secondary" className="ml-auto">{trades.length}</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <div className="max-h-[360px] overflow-y-auto">
          {trades.length === 0 ? (
            <p className="text-center text-muted-foreground text-sm py-8">
              No trades yet
            </p>
          ) : (
            <table className="w-full text-sm">
              <thead className="sticky top-0 bg-card border-b border-border">
                <tr className="text-muted-foreground text-xs">
                  <th className="text-left px-4 py-2">Side</th>
                  <th className="text-left px-4 py-2">Token</th>
                  <th className="text-right px-4 py-2">BNB</th>
                  <th className="text-left px-4 py-2">Tx</th>
                  <th className="text-right px-4 py-2">Status</th>
                  <th className="text-right px-4 py-2">Time</th>
                </tr>
              </thead>
              <tbody>
                {trades.map((t) => (
                  <tr key={t.id} className="border-b border-border/50 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-2">
                      <Badge variant={t.side === "buy" ? "success" : "warning"}>
                        {t.side.toUpperCase()}
                      </Badge>
                    </td>
                    <td className="px-4 py-2 font-mono text-xs">
                      <a
                        href={bscScanLink("token", t.token_address)}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-400 hover:text-blue-300"
                      >
                        {shortenAddress(t.token_address)}
                      </a>
                    </td>
                    <td className="px-4 py-2 text-right font-mono text-xs">
                      {t.amount_bnb?.toFixed(6)}
                    </td>
                    <td className="px-4 py-2 font-mono text-xs">
                      {t.tx_hash ? (
                        <a
                          href={bscScanLink("tx", t.tx_hash)}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1 text-blue-400 hover:text-blue-300"
                        >
                          {shortenHash(t.tx_hash)}
                          <ExternalLink className="h-3 w-3" />
                        </a>
                      ) : (
                        <span className="text-muted-foreground">—</span>
                      )}
                    </td>
                    <td className="px-4 py-2 text-right">
                      <Badge variant={statusVariant(t.status)}>{t.status}</Badge>
                    </td>
                    <td className="px-4 py-2 text-right text-xs text-muted-foreground">
                      {timeAgo(t.executed_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
