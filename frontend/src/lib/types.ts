export interface DailyData {
  date: string;
  cost: number;
  calls: number;
  input_tokens: number;
  output_tokens: number;
}

export interface DailyModelData {
  date: string;
  model: string;
  cost: number;
  calls: number;
  input_tokens: number;
  output_tokens: number;
}

export interface ModelSummary {
  model: string;
  provider: string;
  cost: number;
  calls: number;
  input_tokens: number;
  output_tokens: number;
  avg_latency: number;
}

export interface CallLog {
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

export interface OverviewSummary {
  total_cost: number;
  total_calls: number;
  total_input: number;
  total_output: number;
  avg_latency: number;
}

export interface DashboardData {
  overview: OverviewSummary;
  daily: DailyData[];
  daily_by_model: DailyModelData[];
  models: ModelSummary[];
  model_names: string[];
  calls: CallLog[];
}
