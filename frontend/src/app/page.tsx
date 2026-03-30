"use client";

import { useState, useEffect } from "react";
import { DashboardData } from "@/lib/types";
import { fetchDashboardData, fetchProxyStatus } from "@/lib/api";
import { Header } from "@/components/header";
import { FilterBar } from "@/components/filter-bar";
import { StatCard } from "@/components/stat-card";
import { UsageChart } from "@/components/usage-chart";
import { LogsTable } from "@/components/logs-table";

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [proxyActive, setProxyActive] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Filter state
  const [timeRange, setTimeRange] = useState("30d");
  const [selectedModel, setSelectedModel] = useState("");
  const [groupBy, setGroupBy] = useState<"model" | "total">("model");

  useEffect(() => {
    const controller = new AbortController();
    let interval: ReturnType<typeof setInterval>;

    const load = () => {
      fetchDashboardData(timeRange, selectedModel || undefined, controller.signal)
        .then((d) => {
          if (!controller.signal.aborted) {
            setData(d);
            setError(null);
          }
        })
        .catch((e) => {
          if (!controller.signal.aborted && e.name !== "AbortError") {
            setError(
              "Cannot reach Tirith backend. Is `tirith start` running?"
            );
          }
        });
      fetchProxyStatus(controller.signal).then((active) => {
        if (!controller.signal.aborted) setProxyActive(active);
      });
    };

    load();
    interval = setInterval(load, 60_000);

    return () => {
      controller.abort();
      clearInterval(interval);
    };
  }, [timeRange, selectedModel]);

  if (error) {
    return (
      <div className="min-h-screen bg-[var(--color-bg)]">
        <Header proxyActive={false} />
        <div className="flex flex-col items-center justify-center mt-32 gap-3">
          <div className="text-[var(--color-text-secondary)] text-sm">
            {error}
          </div>
          <code className="text-xs text-[var(--color-text-tertiary)] bg-[var(--color-surface-alt)] px-3 py-1.5 rounded-md border border-[var(--color-border-light)]">
            tirith start
          </code>
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="min-h-screen bg-[var(--color-bg)] flex items-center justify-center">
        <div className="text-[var(--color-text-tertiary)] text-sm">
          Loading...
        </div>
      </div>
    );
  }

  const { overview } = data;

  function formatTokens(n: number): string {
    if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
    if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
    return n.toString();
  }

  return (
    <div className="min-h-screen bg-[var(--color-bg)]">
      <Header proxyActive={proxyActive} />

      <div className="max-w-[1200px] mx-auto px-8 py-6">
        {/* Filter Bar */}
        <div className="mb-6">
          <FilterBar
            models={data.model_names}
            selectedModel={selectedModel}
            onModelChange={setSelectedModel}
            timeRange={timeRange}
            onTimeRangeChange={setTimeRange}
            groupBy={groupBy}
            onGroupByChange={setGroupBy}
          />
        </div>

        {/* Stat Cards */}
        <div className="flex gap-4 mb-6 flex-wrap">
          <StatCard
            label="Total tokens in"
            value={formatTokens(overview.total_input)}
          />
          <StatCard
            label="Total tokens out"
            value={formatTokens(overview.total_output)}
          />
          <StatCard
            label="Total cost"
            value={`$${overview.total_cost.toFixed(2)}`}
          />
          <StatCard
            label="Total calls"
            value={overview.total_calls.toLocaleString()}
          />
        </div>

        {/* Chart */}
        <div className="mb-6">
          <UsageChart
            daily={data.daily}
            dailyByModel={data.daily_by_model}
            groupBy={groupBy}
          />
        </div>

        {/* Logs */}
        <LogsTable calls={data.calls} />
      </div>
    </div>
  );
}
