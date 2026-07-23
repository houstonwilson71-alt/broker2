"use client";
import { useEffect, useState } from "react";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api, type Trade } from "@/lib/api";

interface DataPoint {
  time: string;
  cumPnl: number;
  trades: number;
}

function buildChartData(trades: Trade[]): DataPoint[] {
  const sorted = [...trades]
    .filter((t) => t.status === "confirmed")
    .sort((a, b) => new Date(a.executed_at).getTime() - new Date(b.executed_at).getTime());

  let cum = 0;
  return sorted.map((t, i) => {
    const val = t.side === "sell" ? t.amount_bnb ?? 0 : -(t.amount_bnb ?? 0);
    cum += val;
    return {
      time: new Date(t.executed_at).toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" }),
      cumPnl: parseFloat(cum.toFixed(6)),
      trades: i + 1,
    };
  });
}

interface Props {
  wsEvent: unknown;
}

export function PnLChart({ wsEvent }: Props) {
  const [data, setData] = useState<DataPoint[]>([]);
  const [loading, setLoading] = useState(true);

  async function loadTrades() {
    try {
      const result = await api.trades(200);
      setData(buildChartData(result.data ?? []));
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadTrades();
  }, []);

  useEffect(() => {
    if (!wsEvent) return;
    const ev = wsEvent as { type: string };
    if (ev.type === "sell_executed" || ev.type === "position_opened") {
      loadTrades();
    }
  }, [wsEvent]);

  const lastPnl = data.length > 0 ? data[data.length - 1].cumPnl : 0;
  const isPositive = lastPnl >= 0;
  const color = isPositive ? "#22c55e" : "#ef4444";

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base flex items-center gap-2">
          Cumulative P&L
          <span className={`ml-auto text-lg font-bold ${isPositive ? "text-green-400" : "text-red-400"}`}>
            {isPositive ? "+" : ""}{lastPnl.toFixed(6)} BNB
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="h-[200px] flex items-center justify-center text-muted-foreground text-sm">
            Loading chart…
          </div>
        ) : data.length === 0 ? (
          <div className="h-[200px] flex items-center justify-center text-muted-foreground text-sm">
            No confirmed trades yet
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={data} margin={{ top: 5, right: 5, left: 5, bottom: 5 }}>
              <defs>
                <linearGradient id="pnlGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor={color} stopOpacity={0.3} />
                  <stop offset="95%" stopColor={color} stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="hsl(217 33% 17%)" />
              <XAxis
                dataKey="time"
                tick={{ fill: "hsl(215 20% 55%)", fontSize: 11 }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                tick={{ fill: "hsl(215 20% 55%)", fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v) => v.toFixed(4)}
                width={70}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: "hsl(222 47% 8%)",
                  border: "1px solid hsl(217 33% 17%)",
                  borderRadius: "8px",
                  color: "hsl(210 40% 98%)",
                  fontSize: "12px",
                }}
                formatter={(value: number) => [`${value >= 0 ? "+" : ""}${value.toFixed(6)} BNB`, "Cum. P&L"]}
              />
              <Area
                type="monotone"
                dataKey="cumPnl"
                stroke={color}
                fill="url(#pnlGrad)"
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 4, fill: color }}
              />
            </AreaChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  );
}
