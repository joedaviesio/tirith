# Tirith Dashboard — Charts & Visualizations

## Chart Library

**Recharts** (v3.8.1) — the only charting library used.

## Charts

### 1. Usage Chart — `frontend/src/components/usage-chart.tsx`

Two rendering modes depending on the selected time range:

| Time Range | Chart Type | Description |
|---|---|---|
| **1 hour** | `ComposedChart` (Bar + Line) | Bars show spend per 5-minute bucket. Line overlays a trend. Dual axes: cost (left) + call count (right). |
| **24h / 7d / 30d** | `AreaChart` | Area fill showing daily cumulative spend with a gradient (opacity 0.6 → 0.1). |

**Recharts components used:** `AreaChart`, `Area`, `ComposedChart`, `Bar`, `Line`, `XAxis`, `YAxis`, `CartesianGrid`, `Tooltip`, `ResponsiveContainer`

### 2. Custom Tooltip — `frontend/src/components/custom-tooltip.tsx`

Custom hover tooltip renderer shared by both chart modes. Filters out the "Trend" line entry, displays color dots alongside formatted values.

### 3. Logs Table — `frontend/src/components/logs-table.tsx`

Paginated HTML data table with columns: WHEN, MODEL, COST, SPEED. Not a chart, but the other main data visualization on the dashboard.

## Theming

CSS variables defined in `frontend/src/app/globals.css` (lines 15–19):

| Variable | Value | Usage |
|---|---|---|
| `--chart-line` | `#1b3a2a` | Dark green — primary stroke color |
| `--chart-area-fill` | `#d6ecda` | Light green — area fill |
| `--chart-dot-active` | `#3b3bf9` | Blue — active dot color |
| `--chart-gridline` | `#f0f0f0` | Light gray — gridlines |

Model color palette in `frontend/src/lib/constants.ts`:

| Model | Color |
|---|---|
| Claude Sonnet | `#c4928a` |
| Claude Opus | `#3b3bf9` |
| Claude Haiku | `#8bb8a8` |
| Other | `#d4c5a9` |

## Data Flow

The dashboard page (`frontend/src/app/page.tsx`, lines 119–123) passes `daily`, `calls`, and `timeRange` props to `UsageChart`. Data refreshes every 10 seconds.
