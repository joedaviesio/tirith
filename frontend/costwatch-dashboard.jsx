import { useState, useMemo } from "react";
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Area, AreaChart } from "recharts";

// Mock data matching Go backend JSON structure
const dailySpend = [
  { date: "Mar 18", cost: 4.23, calls: 342 },
  { date: "Mar 19", cost: 6.71, calls: 512 },
  { date: "Mar 20", cost: 8.14, calls: 623 },
  { date: "Mar 21", cost: 5.92, calls: 445 },
  { date: "Mar 22", cost: 12.37, calls: 891 },
  { date: "Mar 23", cost: 15.84, calls: 1102 },
  { date: "Mar 24", cost: 11.20, calls: 834 },
  { date: "Mar 25", cost: 9.56, calls: 701 },
  { date: "Mar 26", cost: 14.22, calls: 987 },
  { date: "Mar 27", cost: 18.41, calls: 1243 },
  { date: "Mar 28", cost: 16.33, calls: 1098 },
  { date: "Mar 29", cost: 13.87, calls: 945 },
  { date: "Mar 30", cost: 7.62, calls: 523 },
];

const tagBreakdown = [
  { tag: "grant-scanner", cost: 48.72, calls: 3201, avg_cost: 0.0152, pct: 33.8 },
  { tag: "chatbot", cost: 35.19, calls: 2845, avg_cost: 0.0124, pct: 24.4 },
  { tag: "summariser", cost: 28.44, calls: 1567, avg_cost: 0.0181, pct: 19.7 },
  { tag: "case-manager", cost: 18.93, calls: 1102, avg_cost: 0.0172, pct: 13.1 },
  { tag: "untagged", cost: 12.94, calls: 531, avg_cost: 0.0244, pct: 9.0 },
];

const modelBreakdown = [
  { model: "claude-sonnet-4-6", cost: 72.41, calls: 5890, input_tokens: 14230000, output_tokens: 3210000, avg_latency: 1240 },
  { model: "claude-haiku-4-5", cost: 18.33, calls: 2890, input_tokens: 8450000, output_tokens: 1890000, avg_latency: 420 },
  { model: "claude-opus-4-6", cost: 42.18, calls: 312, input_tokens: 1890000, output_tokens: 420000, avg_latency: 3200 },
  { model: "gpt-4o-mini", cost: 11.30, calls: 1154, input_tokens: 5670000, output_tokens: 1230000, avg_latency: 680 },
];

const userBreakdown = [
  { user: "user_a1b2c3", cost: 22.14, calls: 1432, first_seen: "Mar 18", last_seen: "Mar 30" },
  { user: "user_d4e5f6", cost: 18.77, calls: 1201, first_seen: "Mar 19", last_seen: "Mar 30" },
  { user: "user_g7h8i9", cost: 15.33, calls: 987, first_seen: "Mar 18", last_seen: "Mar 29" },
  { user: "user_j0k1l2", cost: 14.91, calls: 856, first_seen: "Mar 20", last_seen: "Mar 30" },
  { user: "user_m3n4o5", cost: 12.44, calls: 734, first_seen: "Mar 21", last_seen: "Mar 30" },
  { user: "user_p6q7r8", cost: 11.02, calls: 623, first_seen: "Mar 22", last_seen: "Mar 28" },
  { user: "user_s9t0u1", cost: 9.88, calls: 512, first_seen: "Mar 18", last_seen: "Mar 30" },
  { user: "user_v2w3x4", cost: 8.21, calls: 445, first_seen: "Mar 23", last_seen: "Mar 30" },
  { user: "user_y5z6a7", cost: 7.56, calls: 334, first_seen: "Mar 24", last_seen: "Mar 29" },
  { user: "user_b8c9d0", cost: 5.96, calls: 267, first_seen: "Mar 25", last_seen: "Mar 30" },
];

const PIE_COLORS = ["#22d3ee", "#a78bfa", "#f472b6", "#34d399", "#fbbf24"];
const MODEL_COLORS = { "claude-sonnet-4-6": "#22d3ee", "claude-haiku-4-5": "#34d399", "claude-opus-4-6": "#a78bfa", "gpt-4o-mini": "#fbbf24" };

const tabs = [
  { id: "overview", label: "Overview" },
  { id: "tags", label: "By Feature" },
  { id: "models", label: "By Model" },
  { id: "users", label: "By User" },
];

const CustomTooltip = ({ active, payload, label }) => {
  if (!active || !payload?.length) return null;
  return (
    <div style={{ background: "#1a1a2e", border: "1px solid rgba(255,255,255,0.08)", borderRadius: 8, padding: "10px 14px", fontSize: 13, color: "#e2e8f0", boxShadow: "0 8px 32px rgba(0,0,0,0.4)" }}>
      <div style={{ color: "#94a3b8", marginBottom: 4, fontSize: 12 }}>{label}</div>
      {payload.map((p, i) => (
        <div key={i} style={{ display: "flex", alignItems: "center", gap: 6, marginTop: 2 }}>
          <span style={{ width: 8, height: 8, borderRadius: "50%", background: p.color, display: "inline-block" }} />
          <span style={{ color: "#cbd5e1" }}>{p.name}:</span>
          <span style={{ color: "#fff", fontWeight: 600 }}>{typeof p.value === "number" && p.name !== "Calls" ? `$${p.value.toFixed(2)}` : p.value.toLocaleString()}</span>
        </div>
      ))}
    </div>
  );
};

function StatCard({ label, value, sub, accent }) {
  return (
    <div style={{ background: "rgba(255,255,255,0.03)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, padding: "20px 24px", flex: 1, minWidth: 180 }}>
      <div style={{ fontSize: 12, color: "#64748b", textTransform: "uppercase", letterSpacing: "0.08em", marginBottom: 8 }}>{label}</div>
      <div style={{ fontSize: 28, fontWeight: 700, color: accent || "#f1f5f9", fontFamily: "'JetBrains Mono', 'Fira Code', monospace", letterSpacing: "-0.02em" }}>{value}</div>
      {sub && <div style={{ fontSize: 12, color: "#64748b", marginTop: 6 }}>{sub}</div>}
    </div>
  );
}

function OverviewTab() {
  const totalCost = dailySpend.reduce((s, d) => s + d.cost, 0);
  const totalCalls = dailySpend.reduce((s, d) => s + d.calls, 0);
  const avgDaily = totalCost / dailySpend.length;
  const todayCost = dailySpend[dailySpend.length - 1].cost;

  return (
    <div>
      <div style={{ display: "flex", gap: 16, marginBottom: 32, flexWrap: "wrap" }}>
        <StatCard label="Total Spend (13d)" value={`$${totalCost.toFixed(2)}`} sub={`${totalCalls.toLocaleString()} API calls`} accent="#22d3ee" />
        <StatCard label="Today" value={`$${todayCost.toFixed(2)}`} sub={`${dailySpend[dailySpend.length - 1].calls} calls`} accent="#34d399" />
        <StatCard label="Avg Daily" value={`$${avgDaily.toFixed(2)}`} sub="per day" accent="#a78bfa" />
        <StatCard label="Costliest Feature" value="grant-scanner" sub="$48.72 (33.8%)" accent="#f472b6" />
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 24 }}>
        <div style={{ background: "rgba(255,255,255,0.02)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, padding: 24 }}>
          <div style={{ fontSize: 14, color: "#94a3b8", marginBottom: 16, fontWeight: 500 }}>Daily Spend</div>
          <ResponsiveContainer width="100%" height={240}>
            <AreaChart data={dailySpend}>
              <defs>
                <linearGradient id="costGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#22d3ee" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#22d3ee" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" />
              <XAxis dataKey="date" tick={{ fill: "#64748b", fontSize: 11 }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fill: "#64748b", fontSize: 11 }} axisLine={false} tickLine={false} tickFormatter={v => `$${v}`} />
              <Tooltip content={<CustomTooltip />} />
              <Area type="monotone" dataKey="cost" stroke="#22d3ee" strokeWidth={2} fill="url(#costGrad)" name="Cost" />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        <div style={{ background: "rgba(255,255,255,0.02)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, padding: 24 }}>
          <div style={{ fontSize: 14, color: "#94a3b8", marginBottom: 16, fontWeight: 500 }}>Spend by Feature</div>
          <div style={{ display: "flex", alignItems: "center", gap: 24 }}>
            <ResponsiveContainer width="50%" height={240}>
              <PieChart>
                <Pie data={tagBreakdown} cx="50%" cy="50%" innerRadius={55} outerRadius={85} dataKey="cost" nameKey="tag" stroke="none">
                  {tagBreakdown.map((_, i) => <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />)}
                </Pie>
                <Tooltip content={<CustomTooltip />} />
              </PieChart>
            </ResponsiveContainer>
            <div style={{ flex: 1 }}>
              {tagBreakdown.map((t, i) => (
                <div key={t.tag} style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 10 }}>
                  <span style={{ width: 10, height: 10, borderRadius: 3, background: PIE_COLORS[i], flexShrink: 0 }} />
                  <span style={{ color: "#cbd5e1", fontSize: 13, flex: 1 }}>{t.tag}</span>
                  <span style={{ color: "#f1f5f9", fontSize: 13, fontWeight: 600, fontFamily: "'JetBrains Mono', monospace" }}>${t.cost.toFixed(2)}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function TagsTab() {
  return (
    <div>
      <div style={{ background: "rgba(255,255,255,0.02)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, padding: 24, marginBottom: 24 }}>
        <div style={{ fontSize: 14, color: "#94a3b8", marginBottom: 16, fontWeight: 500 }}>Cost by Feature Tag</div>
        <ResponsiveContainer width="100%" height={280}>
          <BarChart data={tagBreakdown} layout="vertical" margin={{ left: 20 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" horizontal={false} />
            <XAxis type="number" tick={{ fill: "#64748b", fontSize: 11 }} axisLine={false} tickLine={false} tickFormatter={v => `$${v}`} />
            <YAxis type="category" dataKey="tag" tick={{ fill: "#cbd5e1", fontSize: 12 }} axisLine={false} tickLine={false} width={120} />
            <Tooltip content={<CustomTooltip />} />
            <Bar dataKey="cost" name="Cost" radius={[0, 6, 6, 0]} barSize={28}>
              {tagBreakdown.map((_, i) => <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />)}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div style={{ background: "rgba(255,255,255,0.02)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, overflow: "hidden" }}>
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
          <thead>
            <tr style={{ borderBottom: "1px solid rgba(255,255,255,0.06)" }}>
              {["Feature Tag", "Total Cost", "API Calls", "Avg Cost/Call", "% of Spend"].map(h => (
                <th key={h} style={{ padding: "14px 20px", textAlign: "left", color: "#64748b", fontWeight: 500, fontSize: 11, textTransform: "uppercase", letterSpacing: "0.06em" }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {tagBreakdown.map((t, i) => (
              <tr key={t.tag} style={{ borderBottom: "1px solid rgba(255,255,255,0.03)" }}>
                <td style={{ padding: "12px 20px" }}>
                  <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <span style={{ width: 8, height: 8, borderRadius: 2, background: PIE_COLORS[i] }} />
                    <span style={{ color: "#e2e8f0", fontFamily: "'JetBrains Mono', monospace", fontSize: 12 }}>{t.tag}</span>
                  </div>
                </td>
                <td style={{ padding: "12px 20px", color: "#f1f5f9", fontWeight: 600, fontFamily: "'JetBrains Mono', monospace" }}>${t.cost.toFixed(2)}</td>
                <td style={{ padding: "12px 20px", color: "#cbd5e1" }}>{t.calls.toLocaleString()}</td>
                <td style={{ padding: "12px 20px", color: "#cbd5e1", fontFamily: "'JetBrains Mono', monospace" }}>${t.avg_cost.toFixed(4)}</td>
                <td style={{ padding: "12px 20px" }}>
                  <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <div style={{ flex: 1, height: 6, background: "rgba(255,255,255,0.06)", borderRadius: 3, overflow: "hidden", maxWidth: 100 }}>
                      <div style={{ width: `${t.pct}%`, height: "100%", background: PIE_COLORS[i], borderRadius: 3 }} />
                    </div>
                    <span style={{ color: "#94a3b8", fontSize: 12 }}>{t.pct}%</span>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ModelsTab() {
  const totalCost = modelBreakdown.reduce((s, m) => s + m.cost, 0);

  const savings = useMemo(() => {
    const sonnet = modelBreakdown.find(m => m.model === "claude-sonnet-4-6");
    const opus = modelBreakdown.find(m => m.model === "claude-opus-4-6");
    if (!sonnet || !opus) return null;
    const opusInputRate = 15 / 1_000_000;
    const haikuInputRate = 0.8 / 1_000_000;
    const haikuOutputRate = 4 / 1_000_000;
    const opusOutputRate = 75 / 1_000_000;
    const potentialSaving = opus.cost - (opus.input_tokens * haikuInputRate + opus.output_tokens * haikuOutputRate);
    return potentialSaving;
  }, []);

  return (
    <div>
      {savings && (
        <div style={{ background: "rgba(52, 211, 153, 0.08)", border: "1px solid rgba(52, 211, 153, 0.2)", borderRadius: 12, padding: "16px 24px", marginBottom: 24, display: "flex", alignItems: "center", gap: 12 }}>
          <span style={{ fontSize: 20 }}>💡</span>
          <div>
            <div style={{ color: "#34d399", fontSize: 14, fontWeight: 600 }}>Potential savings: ${savings.toFixed(2)}/period</div>
            <div style={{ color: "#94a3b8", fontSize: 12, marginTop: 2 }}>Switching claude-opus-4-6 calls to claude-haiku-4-5 would save {((savings / totalCost) * 100).toFixed(0)}% of total spend</div>
          </div>
        </div>
      )}

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 24, marginBottom: 24 }}>
        <div style={{ background: "rgba(255,255,255,0.02)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, padding: 24 }}>
          <div style={{ fontSize: 14, color: "#94a3b8", marginBottom: 16, fontWeight: 500 }}>Cost by Model</div>
          <ResponsiveContainer width="100%" height={240}>
            <BarChart data={modelBreakdown}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" />
              <XAxis dataKey="model" tick={{ fill: "#64748b", fontSize: 10 }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fill: "#64748b", fontSize: 11 }} axisLine={false} tickLine={false} tickFormatter={v => `$${v}`} />
              <Tooltip content={<CustomTooltip />} />
              <Bar dataKey="cost" name="Cost" radius={[6, 6, 0, 0]} barSize={40}>
                {modelBreakdown.map((m) => <Cell key={m.model} fill={MODEL_COLORS[m.model] || "#64748b"} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div style={{ background: "rgba(255,255,255,0.02)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, padding: 24 }}>
          <div style={{ fontSize: 14, color: "#94a3b8", marginBottom: 16, fontWeight: 500 }}>Avg Latency by Model</div>
          <ResponsiveContainer width="100%" height={240}>
            <BarChart data={modelBreakdown}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" />
              <XAxis dataKey="model" tick={{ fill: "#64748b", fontSize: 10 }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fill: "#64748b", fontSize: 11 }} axisLine={false} tickLine={false} tickFormatter={v => `${v}ms`} />
              <Tooltip content={<CustomTooltip />} />
              <Bar dataKey="avg_latency" name="Latency (ms)" radius={[6, 6, 0, 0]} barSize={40} fill="#a78bfa">
                {modelBreakdown.map((m) => <Cell key={m.model} fill={MODEL_COLORS[m.model] || "#64748b"} opacity={0.7} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div style={{ background: "rgba(255,255,255,0.02)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, overflow: "hidden" }}>
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
          <thead>
            <tr style={{ borderBottom: "1px solid rgba(255,255,255,0.06)" }}>
              {["Model", "Total Cost", "Calls", "Input Tokens", "Output Tokens", "Avg Latency"].map(h => (
                <th key={h} style={{ padding: "14px 20px", textAlign: "left", color: "#64748b", fontWeight: 500, fontSize: 11, textTransform: "uppercase", letterSpacing: "0.06em" }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {modelBreakdown.map(m => (
              <tr key={m.model} style={{ borderBottom: "1px solid rgba(255,255,255,0.03)" }}>
                <td style={{ padding: "12px 20px" }}>
                  <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <span style={{ width: 8, height: 8, borderRadius: "50%", background: MODEL_COLORS[m.model] || "#64748b" }} />
                    <span style={{ color: "#e2e8f0", fontFamily: "'JetBrains Mono', monospace", fontSize: 12 }}>{m.model}</span>
                  </div>
                </td>
                <td style={{ padding: "12px 20px", color: "#f1f5f9", fontWeight: 600, fontFamily: "'JetBrains Mono', monospace" }}>${m.cost.toFixed(2)}</td>
                <td style={{ padding: "12px 20px", color: "#cbd5e1" }}>{m.calls.toLocaleString()}</td>
                <td style={{ padding: "12px 20px", color: "#cbd5e1" }}>{(m.input_tokens / 1_000_000).toFixed(1)}M</td>
                <td style={{ padding: "12px 20px", color: "#cbd5e1" }}>{(m.output_tokens / 1_000_000).toFixed(1)}M</td>
                <td style={{ padding: "12px 20px", color: "#cbd5e1" }}>{m.avg_latency.toLocaleString()}ms</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function UsersTab() {
  const [sortBy, setSortBy] = useState("cost");
  const sorted = useMemo(() => [...userBreakdown].sort((a, b) => b[sortBy] - a[sortBy]), [sortBy]);
  const maxCost = Math.max(...userBreakdown.map(u => u.cost));

  return (
    <div>
      <div style={{ display: "flex", gap: 16, marginBottom: 24, flexWrap: "wrap" }}>
        <StatCard label="Tracked Users" value={userBreakdown.length} accent="#22d3ee" />
        <StatCard label="Highest Cost User" value={`$${userBreakdown[0].cost.toFixed(2)}`} sub={userBreakdown[0].user} accent="#f472b6" />
        <StatCard label="Avg Cost/User" value={`$${(userBreakdown.reduce((s, u) => s + u.cost, 0) / userBreakdown.length).toFixed(2)}`} accent="#a78bfa" />
      </div>

      <div style={{ background: "rgba(255,255,255,0.02)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, overflow: "hidden" }}>
        <div style={{ padding: "16px 20px", borderBottom: "1px solid rgba(255,255,255,0.06)", display: "flex", alignItems: "center", gap: 12 }}>
          <span style={{ color: "#94a3b8", fontSize: 12 }}>Sort by:</span>
          {["cost", "calls"].map(s => (
            <button key={s} onClick={() => setSortBy(s)} style={{ background: sortBy === s ? "rgba(34,211,238,0.15)" : "rgba(255,255,255,0.04)", border: sortBy === s ? "1px solid rgba(34,211,238,0.3)" : "1px solid rgba(255,255,255,0.06)", borderRadius: 6, padding: "4px 12px", color: sortBy === s ? "#22d3ee" : "#94a3b8", fontSize: 12, cursor: "pointer", textTransform: "capitalize", fontWeight: 500 }}>
              {s}
            </button>
          ))}
        </div>
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
          <thead>
            <tr style={{ borderBottom: "1px solid rgba(255,255,255,0.06)" }}>
              {["#", "User", "Total Cost", "API Calls", "Cost/Call", "Spend Bar", "Last Seen"].map(h => (
                <th key={h} style={{ padding: "14px 20px", textAlign: "left", color: "#64748b", fontWeight: 500, fontSize: 11, textTransform: "uppercase", letterSpacing: "0.06em" }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sorted.map((u, i) => (
              <tr key={u.user} style={{ borderBottom: "1px solid rgba(255,255,255,0.03)" }}>
                <td style={{ padding: "12px 20px", color: "#64748b", fontSize: 12 }}>{i + 1}</td>
                <td style={{ padding: "12px 20px" }}>
                  <span style={{ color: "#e2e8f0", fontFamily: "'JetBrains Mono', monospace", fontSize: 12, background: "rgba(255,255,255,0.04)", padding: "2px 8px", borderRadius: 4 }}>{u.user}</span>
                </td>
                <td style={{ padding: "12px 20px", color: "#f1f5f9", fontWeight: 600, fontFamily: "'JetBrains Mono', monospace" }}>${u.cost.toFixed(2)}</td>
                <td style={{ padding: "12px 20px", color: "#cbd5e1" }}>{u.calls.toLocaleString()}</td>
                <td style={{ padding: "12px 20px", color: "#cbd5e1", fontFamily: "'JetBrains Mono', monospace" }}>${(u.cost / u.calls).toFixed(4)}</td>
                <td style={{ padding: "12px 20px" }}>
                  <div style={{ width: 120, height: 6, background: "rgba(255,255,255,0.06)", borderRadius: 3, overflow: "hidden" }}>
                    <div style={{ width: `${(u.cost / maxCost) * 100}%`, height: "100%", background: `hsl(${180 - (u.cost / maxCost) * 120}, 70%, 55%)`, borderRadius: 3, transition: "width 0.5s ease" }} />
                  </div>
                </td>
                <td style={{ padding: "12px 20px", color: "#94a3b8", fontSize: 12 }}>{u.last_seen}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default function CostWatchDashboard() {
  const [activeTab, setActiveTab] = useState("overview");

  return (
    <div style={{ minHeight: "100vh", background: "#0f0f1a", color: "#e2e8f0", fontFamily: "'Inter', -apple-system, sans-serif" }}>
      {/* Header */}
      <div style={{ borderBottom: "1px solid rgba(255,255,255,0.06)", padding: "16px 32px", display: "flex", alignItems: "center", justifyContent: "space-between" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <div style={{ width: 32, height: 32, borderRadius: 8, background: "linear-gradient(135deg, #22d3ee, #a78bfa)", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 16, fontWeight: 700, color: "#0f0f1a" }}>C</div>
          <span style={{ fontSize: 18, fontWeight: 700, letterSpacing: "-0.02em" }}>CostWatch</span>
          <span style={{ fontSize: 11, color: "#64748b", background: "rgba(255,255,255,0.04)", padding: "2px 8px", borderRadius: 4, marginLeft: 4 }}>v0.1.0</span>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
            <span style={{ width: 7, height: 7, borderRadius: "50%", background: "#34d399", boxShadow: "0 0 8px rgba(52,211,153,0.5)" }} />
            <span style={{ fontSize: 12, color: "#94a3b8" }}>Proxy active on :5555</span>
          </div>
          <span style={{ fontSize: 12, color: "#64748b" }}>localhost</span>
        </div>
      </div>

      {/* Tabs */}
      <div style={{ borderBottom: "1px solid rgba(255,255,255,0.06)", padding: "0 32px", display: "flex", gap: 0 }}>
        {tabs.map(tab => (
          <button key={tab.id} onClick={() => setActiveTab(tab.id)} style={{ padding: "14px 24px", background: "none", border: "none", color: activeTab === tab.id ? "#22d3ee" : "#64748b", fontSize: 13, fontWeight: 500, cursor: "pointer", borderBottom: activeTab === tab.id ? "2px solid #22d3ee" : "2px solid transparent", transition: "all 0.2s ease", marginBottom: -1 }}>
            {tab.label}
          </button>
        ))}
      </div>

      {/* Content */}
      <div style={{ padding: "28px 32px", maxWidth: 1200, margin: "0 auto" }}>
        {activeTab === "overview" && <OverviewTab />}
        {activeTab === "tags" && <TagsTab />}
        {activeTab === "models" && <ModelsTab />}
        {activeTab === "users" && <UsersTab />}
      </div>
    </div>
  );
}
