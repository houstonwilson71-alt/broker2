"use client";
import { useEffect, useRef, useState, useCallback } from "react";
import { Bot, RefreshCw } from "lucide-react";
import { BotControl } from "@/components/BotControl";
import { TokenFeed } from "@/components/TokenFeed";
import { TradeLog } from "@/components/TradeLog";
import { Positions } from "@/components/Positions";
import { PnLChart } from "@/components/PnLChart";
import { ConfigPanel } from "@/components/ConfigPanel";
import { StatsBar } from "@/components/StatsBar";
import { api, createWebSocket, type BotStatus, type Position } from "@/lib/api";

const POLL_INTERVAL = 10_000;

export default function Dashboard() {
  const [status, setStatus] = useState<BotStatus | null>(null);
  const [positions, setPositions] = useState<Position[]>([]);
  const [wsEvent, setWsEvent] = useState<unknown>(null);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const refreshStatus = useCallback(async () => {
    try {
      const s = await api.botStatus();
      setStatus(s);
    } catch {
      // backend may not be up yet
    }
  }, []);

  const refreshPositions = useCallback(async () => {
    try {
      const r = await api.positions();
      setPositions(r.data ?? []);
    } catch {
      // ignore
    }
  }, []);

  // Initial load
  useEffect(() => {
    refreshStatus();
    refreshPositions();
  }, [refreshStatus, refreshPositions]);

  // Polling fallback for status
  useEffect(() => {
    pollRef.current = setInterval(() => {
      refreshStatus();
    }, POLL_INTERVAL);
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, [refreshStatus]);

  // WebSocket connection with auto-reconnect
  useEffect(() => {
    let reconnectTimeout: ReturnType<typeof setTimeout>;

    function connect() {
      try {
        const ws = createWebSocket((data) => {
          setWsEvent(data);
          // Refresh status on relevant events
          const ev = data as { type?: string };
          if (
            ev.type === "position_opened" ||
            ev.type === "position_opened_simulated" ||
            ev.type === "sell_executed"
          ) {
            refreshStatus();
            refreshPositions();
          }
        });

        ws.onopen = () => setConnected(true);
        ws.onclose = () => {
          setConnected(false);
          reconnectTimeout = setTimeout(connect, 3000);
        };
        ws.onerror = () => {
          ws.close();
        };
        wsRef.current = ws;
      } catch {
        reconnectTimeout = setTimeout(connect, 5000);
      }
    }

    connect();
    return () => {
      clearTimeout(reconnectTimeout);
      wsRef.current?.close();
    };
  }, [refreshStatus, refreshPositions]);

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border bg-card/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-screen-2xl mx-auto px-4 sm:px-6 h-14 flex items-center gap-4">
          <div className="flex items-center gap-2.5">
            <div className="h-7 w-7 rounded-lg bg-primary/20 border border-primary/30 flex items-center justify-center">
              <Bot className="h-4 w-4 text-primary" />
            </div>
            <span className="font-bold text-sm tracking-tight">BSC Sniper</span>
          </div>

          <div className="h-5 w-px bg-border mx-1" />

          <div className="flex-1">
            <BotControl status={status} onStatusChange={refreshStatus} />
          </div>

          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span
              className={`h-1.5 w-1.5 rounded-full ${connected ? "bg-green-500" : "bg-red-500"}`}
            />
            <span>{connected ? "WS connected" : "reconnecting…"}</span>
            <button
              onClick={() => { refreshStatus(); refreshPositions(); }}
              className="ml-1 hover:text-foreground transition-colors"
              title="Refresh"
            >
              <RefreshCw className="h-3 w-3" />
            </button>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-screen-2xl mx-auto px-4 sm:px-6 py-6 space-y-6">
        {/* Stats Bar */}
        <StatsBar status={status} positions={positions} />

        {/* P&L Chart */}
        <PnLChart wsEvent={wsEvent} />

        {/* Positions + Config */}
        <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
          <div className="xl:col-span-2">
            <Positions wsEvent={wsEvent} />
          </div>
          <div>
            <ConfigPanel />
          </div>
        </div>

        {/* Token Feed + Trade Log */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
          <TokenFeed wsEvent={wsEvent} />
          <TradeLog wsEvent={wsEvent} />
        </div>
      </main>
    </div>
  );
}
