"use client";
import { Activity, TrendingUp, Zap, BarChart2 } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import type { BotStatus, Position } from "@/lib/api";

interface Props {
  status: BotStatus | null;
  positions: Position[];
}

function StatCard({ icon: Icon, label, value, sub, color }: {
  icon: React.ElementType;
  label: string;
  value: string;
  sub?: string;
  color?: string;
}) {
  return (
    <Card>
      <CardContent className="pt-5 pb-4">
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-lg bg-muted ${color ?? "text-muted-foreground"}`}>
            <Icon className="h-4 w-4" />
          </div>
          <div>
            <p className="text-xs text-muted-foreground">{label}</p>
            <p className="text-xl font-bold leading-none mt-0.5">{value}</p>
            {sub && <p className="text-xs text-muted-foreground mt-0.5">{sub}</p>}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function StatsBar({ status, positions }: Props) {
  const openPositions = positions.filter((p) => p.status === "open" || p.status === "partial");
  const costBNB = openPositions.reduce((s, p) => s + (p.cost_bnb ?? 0), 0);
  const realizedPnl = positions.reduce((s, p) => s + (p.realized_pnl_bnb ?? 0), 0);
  const pnlPositive = realizedPnl >= 0;

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard
        icon={Activity}
        label="Pairs Seen"
        value={(status?.pairs_seen ?? 0).toLocaleString()}
        sub={`${status?.pairs_passed ?? 0} passed filters`}
        color="text-blue-400"
      />
      <StatCard
        icon={Zap}
        label="Total Trades"
        value={(status?.trades_total ?? 0).toLocaleString()}
        sub={`${openPositions.length} open positions`}
        color="text-yellow-400"
      />
      <StatCard
        icon={BarChart2}
        label="Deployed BNB"
        value={costBNB.toFixed(4)}
        sub={`across ${openPositions.length} positions`}
        color="text-purple-400"
      />
      <StatCard
        icon={TrendingUp}
        label="Realized P&L"
        value={`${pnlPositive ? "+" : ""}${realizedPnl.toFixed(6)} BNB`}
        color={pnlPositive ? "text-green-400" : "text-red-400"}
      />
    </div>
  );
}
