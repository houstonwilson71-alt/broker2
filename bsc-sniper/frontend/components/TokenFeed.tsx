"use client";
import { useEffect, useState } from "react";
import { ExternalLink } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { api, type Token } from "@/lib/api";
import { bscScanLink, shortenAddress, timeAgo } from "@/lib/utils";

interface Props {
  wsEvent: unknown;
}

export function TokenFeed({ wsEvent }: Props) {
  const [tokens, setTokens] = useState<Token[]>([]);

  useEffect(() => {
    api.tokens(30).then((r) => setTokens(r.data ?? [])).catch(console.error);
  }, []);

  useEffect(() => {
    if (!wsEvent) return;
    const ev = wsEvent as { type: string; token?: Token };
    if (ev.type === "token_approved" && ev.token) {
      setTokens((prev) => [ev.token!, ...prev].slice(0, 100));
    }
  }, [wsEvent]);

  return (
    <Card className="h-full">
      <CardHeader className="pb-3">
        <CardTitle className="text-base flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-green-500 animate-pulse" />
          Live Token Feed
          <Badge variant="secondary" className="ml-auto">{tokens.length}</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <div className="max-h-[420px] overflow-y-auto">
          {tokens.length === 0 ? (
            <p className="text-center text-muted-foreground text-sm py-8">
              Waiting for new pairs…
            </p>
          ) : (
            <table className="w-full text-sm">
              <thead className="sticky top-0 bg-card border-b border-border">
                <tr className="text-muted-foreground text-xs">
                  <th className="text-left px-4 py-2">Token</th>
                  <th className="text-left px-4 py-2">Address</th>
                  <th className="text-left px-4 py-2">Pair</th>
                  <th className="text-right px-4 py-2">Age</th>
                </tr>
              </thead>
              <tbody>
                {tokens.map((t, i) => (
                  <tr
                    key={t.id ?? i}
                    className="border-b border-border/50 hover:bg-muted/30 transition-colors"
                  >
                    <td className="px-4 py-2">
                      <div className="font-medium text-foreground">{t.symbol || "???"}</div>
                      <div className="text-xs text-muted-foreground">{t.name || "Unknown"}</div>
                    </td>
                    <td className="px-4 py-2 font-mono text-xs">
                      <a
                        href={bscScanLink("token", t.address)}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-1 text-blue-400 hover:text-blue-300"
                      >
                        {shortenAddress(t.address)}
                        <ExternalLink className="h-3 w-3" />
                      </a>
                    </td>
                    <td className="px-4 py-2 font-mono text-xs text-muted-foreground">
                      <a
                        href={bscScanLink("address", t.pair_address)}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-1 hover:text-foreground"
                      >
                        {shortenAddress(t.pair_address)}
                        <ExternalLink className="h-3 w-3" />
                      </a>
                    </td>
                    <td className="px-4 py-2 text-right text-xs text-muted-foreground">
                      {timeAgo(t.created_at)}
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
