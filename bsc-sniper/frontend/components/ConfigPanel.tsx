"use client";
import { useEffect, useState } from "react";
import { Save } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { api, type BotConfig } from "@/lib/api";

export function ConfigPanel() {
  const [cfg, setCfg] = useState<BotConfig | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    api.getConfig().then(setCfg).catch(console.error);
  }, []);

  async function save() {
    if (!cfg) return;
    setSaving(true);
    try {
      const updated = await api.updateConfig(cfg);
      setCfg(updated);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (e) {
      console.error(e);
    } finally {
      setSaving(false);
    }
  }

  if (!cfg) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground text-sm">
          Loading config…
        </CardContent>
      </Card>
    );
  }

  function field(
    label: string,
    key: keyof BotConfig,
    type: "number" | "text" = "number",
    hint?: string
  ) {
    return (
      <div className="space-y-1.5">
        <Label htmlFor={key} className="text-xs text-muted-foreground">
          {label}
          {hint && <span className="ml-1 opacity-60">({hint})</span>}
        </Label>
        <Input
          id={key}
          type={type}
          value={String(cfg![key])}
          onChange={(e) =>
            setCfg((prev) =>
              prev ? { ...prev, [key]: type === "number" ? parseFloat(e.target.value) || 0 : e.target.value } : prev
            )
          }
          className="h-8 text-sm"
        />
      </div>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base flex items-center gap-2">
          Bot Configuration
          <Button
            size="sm"
            onClick={save}
            disabled={saving}
            className="ml-auto h-7 text-xs"
            variant={saved ? "secondary" : "default"}
          >
            <Save className="mr-1 h-3 w-3" />
            {saved ? "Saved!" : saving ? "Saving…" : "Save"}
          </Button>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-5">
        {/* Live trading toggle */}
        <div className="flex items-center justify-between">
          <div>
            <Label className="text-sm font-medium">Live Trading</Label>
            <p className="text-xs text-muted-foreground mt-0.5">
              {cfg.live_trading_enabled
                ? "⚠️ Real transactions will be submitted"
                : "Simulation mode — no real trades"}
            </p>
          </div>
          <Switch
            checked={cfg.live_trading_enabled}
            onCheckedChange={(v) => setCfg((prev) => prev ? { ...prev, live_trading_enabled: v } : prev)}
          />
        </div>

        <Separator />

        <div className="grid grid-cols-2 gap-4">
          {field("Buy Amount (BNB)", "buy_amount_bnb", "number", "per trade")}
          {field("Slippage (BPS)", "slippage_bps", "number", "150 = 1.5%")}
        </div>

        <Separator />
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Safety Filters</p>
        <div className="grid grid-cols-2 gap-4">
          {field("Min Liquidity (USD)", "min_liquidity_usd")}
          {field("Max Pair Age (sec)", "max_age_sec")}
          {field("Min Holders", "min_holders")}
          {field("Max Top-10 % ", "max_top10_pct", "number", "concentration")}
        </div>

        <Separator />
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Exit Strategy</p>
        <div className="grid grid-cols-2 gap-4">
          {field("Take-Profit 1 (%)", "take_profit_1_pct", "number", "sell 50%")}
          {field("Trailing Stop (%)", "trailing_stop_pct", "number", "after TP1")}
        </div>
      </CardContent>
    </Card>
  );
}
