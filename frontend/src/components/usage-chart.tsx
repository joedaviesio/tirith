"use client";

import { useMemo, useState } from "react";
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { DailyData, DailyModelData } from "@/lib/types";
import { getModelColor, getModelShortName } from "@/lib/constants";
import { CustomTooltip } from "./custom-tooltip";

type ChartView = "tokens" | "cost";

interface Props {
  daily: DailyData[];
  dailyByModel: DailyModelData[];
  groupBy: "model" | "total";
}

export function UsageChart({ daily, dailyByModel, groupBy }: Props) {
  const [view, setView] = useState<ChartView>("tokens");

  // Pivot daily-by-model data into stacked format: { date, model1: value, model2: value, ... }
  const { stackedData, modelKeys } = useMemo(() => {
    const byDate: Record<string, Record<string, number>> = {};
    const allModels = new Set<string>();

    for (const d of dailyByModel) {
      if (!byDate[d.date]) byDate[d.date] = {};
      allModels.add(d.model);
      const val =
        view === "cost"
          ? d.cost
          : d.input_tokens + d.output_tokens;
      byDate[d.date][d.model] = (byDate[d.date][d.model] || 0) + val;
    }

    const modelKeys = Array.from(allModels);
    const stackedData = Object.entries(byDate).map(([date, models]) => ({
      date,
      ...models,
    }));

    return { stackedData, modelKeys };
  }, [dailyByModel, view]);

  const formatValue = (v: number) => {
    if (view === "cost") return `$${v.toFixed(2)}`;
    if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(1)}M`;
    if (v >= 1_000) return `${(v / 1_000).toFixed(0)}K`;
    return v.toString();
  };

  const tooltipFormatter = (value: number, name: string) => {
    if (view === "cost") return `$${value.toFixed(2)}`;
    return value.toLocaleString() + " tokens";
  };

  if (daily.length === 0) {
    return (
      <div className="rounded-xl border border-[var(--color-border-light)] bg-[var(--color-surface)] p-6 shadow-[var(--shadow-card)]">
        <div className="flex items-center justify-center py-16 text-[var(--color-text-tertiary)] text-sm">
          No data yet. Make API calls through the proxy to see usage here.
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-xl border border-[var(--color-border-light)] bg-[var(--color-surface)] p-6 shadow-[var(--shadow-card)]">
      {/* Header with toggle */}
      <div className="flex items-center justify-between mb-1">
        <div className="text-[15px] font-semibold text-[var(--color-text-primary)]">
          {view === "tokens" ? "Token usage" : "Daily token cost"}
        </div>
        <div className="flex items-center gap-1">
          <span className="text-[13px] text-[var(--color-text-secondary)] mr-2">
            Choose view
          </span>
          {(
            [
              { key: "tokens", label: "Token usage" },
              { key: "cost", label: "Cost" },
            ] as const
          ).map((t) => (
            <button
              key={t.key}
              onClick={() => setView(t.key)}
              className="px-3 py-1.5 rounded-full text-[13px] font-medium cursor-pointer transition-all"
              style={{
                border:
                  view === t.key
                    ? "1.5px solid var(--color-border-active)"
                    : "1.5px solid var(--color-border-light)",
                background:
                  view === t.key
                    ? "var(--color-surface)"
                    : "var(--color-surface-alt)",
                color: "var(--color-text-primary)",
              }}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>
      <div className="text-[13px] text-[var(--color-text-tertiary)] mb-4">
        Includes usage from all API calls through the proxy
      </div>

      <ResponsiveContainer width="100%" height={320}>
        {groupBy === "model" && stackedData.length > 0 ? (
          <BarChart data={stackedData}>
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="var(--chart-gridline)"
              horizontal={true}
              vertical={false}
            />
            <XAxis
              dataKey="date"
              tick={{ fill: "#666666", fontSize: 12 }}
              axisLine={false}
              tickLine={false}
            />
            <YAxis
              tick={{ fill: "#666666", fontSize: 12 }}
              axisLine={false}
              tickLine={false}
              tickFormatter={formatValue}
            />
            <Tooltip
              content={
                <CustomTooltip formatter={tooltipFormatter} />
              }
            />
            <Legend
              formatter={(value: string) => getModelShortName(value)}
              wrapperStyle={{ fontSize: 12, color: "#666666" }}
            />
            {modelKeys.map((model) => (
              <Bar
                key={model}
                dataKey={model}
                name={model}
                stackId="a"
                fill={getModelColor(model)}
                radius={[0, 0, 0, 0]}
              />
            ))}
          </BarChart>
        ) : (
          <AreaChart data={daily}>
            <defs>
              <linearGradient id="areaFill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#d6ecda" stopOpacity={0.6} />
                <stop offset="100%" stopColor="#d6ecda" stopOpacity={0.1} />
              </linearGradient>
            </defs>
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="var(--chart-gridline)"
              horizontal={true}
              vertical={false}
            />
            <XAxis
              dataKey="date"
              tick={{ fill: "#666666", fontSize: 12 }}
              axisLine={false}
              tickLine={false}
            />
            <YAxis
              tick={{ fill: "#666666", fontSize: 12 }}
              axisLine={false}
              tickLine={false}
              tickFormatter={formatValue}
            />
            <Tooltip
              content={
                <CustomTooltip formatter={tooltipFormatter} />
              }
            />
            {view === "tokens" ? (
              <Area
                type="monotone"
                dataKey="input_tokens"
                stroke="var(--chart-line)"
                strokeWidth={2}
                fill="url(#areaFill)"
                name="Tokens"
              />
            ) : (
              <Area
                type="monotone"
                dataKey="cost"
                stroke="var(--chart-line)"
                strokeWidth={2}
                fill="url(#areaFill)"
                name="Cost"
              />
            )}
          </AreaChart>
        )}
      </ResponsiveContainer>

      {/* Footnote */}
      <div className="mt-4 pt-3 border-t border-[var(--color-border)] flex items-center gap-2 text-[13px] text-[var(--color-text-tertiary)]">
        <span>&#9432;</span>
        <span>
          {view === "tokens"
            ? "Shows total input tokens per day"
            : "Shows total cost in USD per day"}
        </span>
      </div>
    </div>
  );
}
