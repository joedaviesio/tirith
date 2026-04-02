"use client";

import { useMemo } from "react";
import {
  AreaChart,
  Area,
  ComposedChart,
  Bar,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { DailyData, CallLog } from "@/lib/types";
import { CustomTooltip } from "./custom-tooltip";

interface Props {
  daily: DailyData[];
  calls: CallLog[];
  timeRange: string;
}

interface BucketData {
  label: string;
  cost: number;
  calls: number;
}

export function UsageChart({ daily, calls, timeRange }: Props) {
  // For "1h" view, bucket calls into 5-minute intervals
  const hourlyBuckets = useMemo((): BucketData[] => {
    if (timeRange !== "1h") return [];

    const now = new Date();
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    const buckets: BucketData[] = [];

    // Create 12 five-minute buckets
    for (let i = 0; i < 12; i++) {
      const bucketStart = new Date(oneHourAgo.getTime() + i * 5 * 60 * 1000);
      buckets.push({
        label: bucketStart.toLocaleTimeString("en-US", {
          hour: "numeric",
          minute: "2-digit",
          hour12: true,
        }),
        cost: 0,
        calls: 0,
      });
    }

    // Assign calls to buckets
    for (const call of calls) {
      const callTime = new Date(call.timestamp);
      if (callTime < oneHourAgo) continue;
      const minutesSinceStart =
        (callTime.getTime() - oneHourAgo.getTime()) / (1000 * 60);
      const bucketIndex = Math.min(11, Math.floor(minutesSinceStart / 5));
      buckets[bucketIndex].cost += call.cost_cents / 100;
      buckets[bucketIndex].calls += 1;
    }

    return buckets;
  }, [calls, timeRange]);

  const chartData = timeRange === "1h" ? hourlyBuckets : daily;
  const dataKey = timeRange === "1h" ? "label" : "date";

  const formatValue = (v: number) => {
    if (v >= 0.01) return `$${v.toFixed(2)}`;
    if (v > 0) return `$${v.toFixed(4)}`;
    return "$0";
  };

  const tooltipFormatter = (value: number, name: string) => {
    if (name === "calls" || name === "Calls") return `${value}`;
    if (value >= 0.01) return `$${value.toFixed(2)}`;
    return `$${value.toFixed(4)}`;
  };

  const isEmpty =
    timeRange === "1h"
      ? hourlyBuckets.every((b) => b.calls === 0)
      : daily.length === 0;

  if (isEmpty) {
    return (
      <div className="rounded-xl border border-[var(--color-border-light)] bg-[var(--color-surface)] p-6 shadow-[var(--shadow-card)]">
        <div className="flex items-center justify-center py-16 text-[var(--color-text-tertiary)] text-sm">
          No activity yet. Make API calls through the proxy to see spend here.
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-xl border border-[var(--color-border-light)] bg-[var(--color-surface)] p-6 shadow-[var(--shadow-card)]">
      <div className="flex items-center justify-between mb-1">
        <div className="text-[15px] font-semibold text-[var(--color-text-primary)]">
          Spend over time
        </div>
        <div className="text-[13px] text-[var(--color-text-tertiary)]">
          {timeRange === "1h"
            ? "5-minute intervals"
            : "Daily totals"}
        </div>
      </div>
      <div className="text-[13px] text-[var(--color-text-tertiary)] mb-4">
        Cost in USD for all API calls through the proxy
      </div>

      <ResponsiveContainer width="100%" height={280}>
        {timeRange === "1h" ? (
          <ComposedChart data={hourlyBuckets}>
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="var(--chart-gridline)"
              horizontal={true}
              vertical={false}
            />
            <XAxis
              dataKey="label"
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
              content={<CustomTooltip formatter={tooltipFormatter} />}
            />
            <Bar
              dataKey="cost"
              name="Spend"
              fill="var(--chart-line)"
              radius={[4, 4, 0, 0]}
              opacity={0.85}
            />
            <Line
              dataKey="cost"
              name="Trend"
              stroke="var(--color-text-secondary)"
              strokeWidth={1.5}
              dot={{ r: 3.5, fill: "var(--color-text-secondary)", stroke: "var(--color-surface)", strokeWidth: 2 }}
              activeDot={false}
              legendType="none"
            />
          </ComposedChart>
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
              content={<CustomTooltip formatter={tooltipFormatter} />}
            />
            <Area
              type="monotone"
              dataKey="cost"
              stroke="var(--chart-line)"
              strokeWidth={2}
              fill="url(#areaFill)"
              name="Spend"
            />
          </AreaChart>
        )}
      </ResponsiveContainer>
    </div>
  );
}
