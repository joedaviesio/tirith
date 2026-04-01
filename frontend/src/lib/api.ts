import {
  DashboardData,
  DailyData,
  DailyModelData,
  ModelSummary,
  CallLog,
  OverviewSummary,
} from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

async function fetchJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, { cache: "no-store", signal });
  if (!res.ok) throw new Error(`${path}: ${res.status}`);
  return res.json();
}

interface RawOverview {
  total_cost_cents: number;
  total_calls: number;
  total_input: number;
  total_output: number;
  avg_latency_ms: number;
}

interface RawDaily {
  date: string;
  total_cost_cents: number;
  total_calls: number;
  total_input: number;
  total_output: number;
}

interface RawDailyModel {
  date: string;
  model: string;
  total_cost_cents: number;
  total_calls: number;
  total_input: number;
  total_output: number;
}

interface RawModel {
  model: string;
  provider: string;
  total_cost_cents: number;
  total_calls: number;
  total_input: number;
  total_output: number;
  avg_latency_ms: number;
}

interface RawCall {
  id: string;
  timestamp: string;
  provider: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  cost_cents: number;
  latency_ms: number;
  status_code: number;
  streaming: boolean;
  tag: string;
  user_tag: string;
  environment: string;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr + "T00:00:00");
  return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}

export async function fetchDashboardData(
  timeRange: string = "30d",
  model?: string,
  signal?: AbortSignal
): Promise<DashboardData> {
  const params = new URLSearchParams({ last: timeRange });
  if (model) params.set("model", model);
  const qs = params.toString();

  const [rawOverview, rawDaily, rawDailyModel, rawModels, modelNames, rawCalls] =
    await Promise.all([
      fetchJSON<RawOverview>(`/api/overview?${qs}`, signal),
      fetchJSON<RawDaily[]>(`/api/daily-spend?${qs}`, signal),
      fetchJSON<RawDailyModel[]>(`/api/daily-by-model?${qs}`, signal),
      fetchJSON<RawModel[]>(`/api/by-model?${qs}`, signal),
      fetchJSON<string[]>("/api/models?last=90d", signal),
      fetchJSON<RawCall[]>("/api/calls", signal),
    ]);

  const overview: OverviewSummary = {
    total_cost: rawOverview.total_cost_cents / 100,
    total_calls: rawOverview.total_calls,
    total_input: rawOverview.total_input,
    total_output: rawOverview.total_output,
    avg_latency: rawOverview.avg_latency_ms,
  };

  const daily: DailyData[] = (rawDaily || []).map((d) => ({
    date: formatDate(d.date),
    cost: d.total_cost_cents / 100,
    calls: d.total_calls,
    input_tokens: d.total_input,
    output_tokens: d.total_output,
  }));

  const daily_by_model: DailyModelData[] = (rawDailyModel || []).map((d) => ({
    date: formatDate(d.date),
    model: d.model,
    cost: d.total_cost_cents / 100,
    calls: d.total_calls,
    input_tokens: d.total_input,
    output_tokens: d.total_output,
  }));

  const models: ModelSummary[] = (rawModels || []).map((m) => ({
    model: m.model,
    provider: m.provider,
    cost: m.total_cost_cents / 100,
    calls: m.total_calls,
    input_tokens: m.total_input,
    output_tokens: m.total_output,
    avg_latency: m.avg_latency_ms,
  }));

  const calls: CallLog[] = (rawCalls || []).map((c) => ({ ...c }));

  return {
    overview,
    daily,
    daily_by_model,
    models,
    model_names: modelNames || [],
    calls,
  };
}

export async function fetchProxyStatus(signal?: AbortSignal): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/api/proxy-health`, {
      cache: "no-store",
      signal,
    });
    if (!res.ok) return false;
    const data = await res.json();
    return data.active === true;
  } catch {
    return false;
  }
}
