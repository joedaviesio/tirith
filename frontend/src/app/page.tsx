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

  // Filter state — default to last hour for real-time focus
  const [timeRange, setTimeRange] = useState("1h");
  const [selectedModel, setSelectedModel] = useState("");

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
    // Refresh every 10s for near real-time feel
    interval = setInterval(load, 10_000);

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
          />
        </div>

        {/* Stat Cards — just spend and calls, kept simple */}
        <div className="flex gap-4 mb-6 flex-wrap">
          <StatCard
            label="Total spend"
            value={`$${overview.total_cost.toFixed(2)}`}
            sub={`${overview.total_input.toLocaleString()} tokens in, ${overview.total_output.toLocaleString()} out`}
          />
          <StatCard
            label="API calls"
            value={overview.total_calls.toLocaleString()}
            sub={
              overview.avg_latency > 0
                ? `${overview.avg_latency < 1000 ? `${Math.round(overview.avg_latency)}ms` : `${(overview.avg_latency / 1000).toFixed(1)}s`} avg response time`
                : undefined
            }
          />
        </div>

        {/* Chart */}
        <div className="mb-6">
          <UsageChart
            daily={data.daily}
            calls={data.calls}
            timeRange={timeRange}
          />
        </div>

        {/* Recent Calls */}
        <LogsTable calls={data.calls} />
      </div>
    </div>
  );
}
