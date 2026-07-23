"use client";
import { useState } from "react";
import { Play, Square, Activity, Zap } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { api, type BotStatus } from "@/lib/api";
import { timeAgo } from "@/lib/utils";

interface Props {
  status: BotStatus | null;
  onStatusChange: () => void;
}

export function BotControl({ status, onStatusChange }: Props) {
  const [loading, setLoading] = useState(false);

  async function toggle() {
    setLoading(true);
    try {
      if (status?.running) {
        await api.botStop();
      } else {
        await api.botStart();
      }
      onStatusChange();
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }

  const running = status?.running ?? false;

  return (
    <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
      <div className="flex items-center gap-3">
        <div className={`relative flex h-3 w-3`}>
          <span
            className={`absolute inline-flex h-full w-full rounded-full opacity-75 ${
              running ? "animate-ping bg-green-400" : "bg-muted-foreground"
            }`}
          />
          <span
            className={`relative inline-flex rounded-full h-3 w-3 ${
              running ? "bg-green-500" : "bg-muted-foreground"
            }`}
          />
        </div>
        <span className={`font-semibold text-lg ${running ? "text-green-400" : "text-muted-foreground"}`}>
          {running ? "BOT ACTIVE" : "BOT STOPPED"}
        </span>
        {running && status?.started_at && (
          <Badge variant="success">running {timeAgo(status.started_at)}</Badge>
        )}
      </div>

      <Button
        onClick={toggle}
        disabled={loading}
        variant={running ? "destructive" : "default"}
        size="lg"
        className={running ? "" : "bot-running"}
      >
        {loading ? (
          <Activity className="mr-2 h-4 w-4 animate-spin" />
        ) : running ? (
          <Square className="mr-2 h-4 w-4" />
        ) : (
          <Zap className="mr-2 h-4 w-4" />
        )}
        {loading ? "Processing…" : running ? "Stop Bot" : "Start Bot"}
      </Button>

      {status && (
        <div className="flex gap-4 text-sm text-muted-foreground ml-auto">
          <span><span className="text-foreground font-medium">{status.pairs_seen.toLocaleString()}</span> pairs seen</span>
          <span><span className="text-green-400 font-medium">{status.pairs_passed.toLocaleString()}</span> passed</span>
          <span><span className="text-blue-400 font-medium">{status.trades_total.toLocaleString()}</span> trades</span>
        </div>
      )}
    </div>
  );
}
