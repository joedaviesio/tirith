# CostWatch — Project Status

## What's Built

### Go CLI + Proxy (compiles and runs)

The core binary is built. You can run it from this directory with `./costwatch`.

| Command | Status | What it does |
|---------|--------|-------------|
| `./costwatch start` | Working | Starts the proxy on localhost:5555. Accepts API calls, forwards them to Anthropic, logs token counts + cost to SQLite. |
| `./costwatch report` | Working | Reads the SQLite database and prints a cost summary table — total spend, breakdown by model, breakdown by tag. Currently shows $0.00 because no calls have been made yet. |
| `./costwatch run -- <your command>` | Working | Starts the proxy, sets the right env vars, runs your command, then shuts down the proxy when done. |
| `./costwatch stop` | Partial | Tells you how to kill the proxy. No PID file or signal-based stop yet. |

### What the proxy actually does

When your app sends an API call to `localhost:5555` instead of `api.anthropic.com`, the proxy:

1. Reads any `X-CostWatch-Tag` / `X-CostWatch-User` headers (for labelling)
2. Strips those headers so Anthropic never sees them
3. Forwards the request to Anthropic exactly as-is
4. Reads the response — pulls out token counts from the `usage` field
5. Calculates the dollar cost using embedded pricing data
6. Logs everything to a local SQLite database (`~/.costwatch/data.db`)
7. Returns the original response to your app, unmodified

Your app doesn't know the proxy exists. It gets the same response it would from Anthropic directly.

Supports both regular (wait for full response) and streaming (SSE) calls.

### SDK Wrappers (code written, not yet tested)

Small packages that make setup even easier — instead of changing env vars, you just add one import line.

- **Python** (`sdks/python/`) — `import costwatch` at the top of your app, and all Anthropic/OpenAI client instances automatically route through the proxy.
- **TypeScript** (`sdks/typescript/`) — `import "costwatch"` does the same for Node.js apps.

These are not published to PyPI/npm yet. For now they'd be installed locally with `pip install -e ./sdks/python`.

### Pricing Engine

Embedded YAML with current pricing for:
- Anthropic: Opus 4.6, Sonnet 4.6, Haiku 4.5
- OpenAI: GPT-4o, GPT-4o-mini, GPT-4.1, o3, o4-mini

Calculates cost in cents (integers, not floats) to avoid rounding issues.

### Local Storage

SQLite database at `~/.costwatch/data.db`. Stores every API call with: timestamp, provider, model, token counts (input/output/cache), cost, latency, status code, tags, and environment label.

---

## What's NOT Built Yet

### Dashboard (the frontend)

There is no web UI yet. The `internal/dashboard/` directory exists but is empty. This is the piece that would give you charts and visualisations at `localhost:5556`.

**What it needs:**
- A small React app (or even plain HTML + Chart.js for MVP)
- An API layer in the Go binary that serves JSON data from SQLite
- Embed the built frontend into the Go binary so it ships as one file
- Charts: spend over time, by model, by tag, request volume

### End-to-end testing

We haven't sent a real API call through the proxy yet. Need to:
1. Start the proxy
2. Make a real Anthropic API call pointed at `localhost:5555`
3. Confirm the response comes back correctly
4. Run `./costwatch report` and see the cost show up

### PID-based stop

`./costwatch stop` should write/read a PID file so it can actually stop the proxy, instead of telling you to Ctrl+C.

### Global install

Right now the binary only exists in this directory. For it to work as `costwatch` from anywhere, it needs to be either:
- Copied to `/usr/local/bin/`
- Installed via `go install`
- Published via Homebrew or npm

### SDK wrapper testing

The Python and TypeScript wrappers need real-world testing — import them in a project, make API calls, confirm routing works and fallback behaviour is correct.

---

## Steps to Finish the Local Prototype

Here's the path from where we are now to "I can use this and see a working dashboard":

### Step 1: End-to-end test (10 min)

Start the proxy, send a real API call through it (via curl or a Python script), confirm cost shows up in the report.

### Step 2: Dashboard API endpoints (30 min)

Add JSON API routes to the Go binary:
- `GET /api/overview` — total spend, call count, date range
- `GET /api/by-model` — cost breakdown by model
- `GET /api/by-tag` — cost breakdown by tag
- `GET /api/calls` — paginated list of recent calls

These read from the same SQLite database the proxy writes to.

### Step 3: Dashboard frontend (1-2 hours)

Build a simple single-page app. Minimal viable version:
- Summary cards (total spend, total calls, avg latency)
- Line chart: spend over time
- Bar chart: cost by model
- Bar chart: cost by tag
- Table: recent API calls

Options:
- **Simple**: Plain HTML + vanilla JS + Chart.js (no build step, embed directly)
- **Polished**: React + Recharts (needs a build step, then embed the dist)

### Step 4: Embed and serve dashboard (30 min)

Use Go's `embed` package to bundle the frontend files into the binary. Serve them from `localhost:5556` when you run `./costwatch dashboard` or automatically alongside the proxy.

### Step 5: Python SDK wrapper test (15 min)

Install the wrapper locally (`pip install -e ./sdks/python`), write a small test script that imports `costwatch` then uses the Anthropic SDK normally, confirm calls route through the proxy.

### Step 6: Polish (30 min)

- PID file for `costwatch stop`
- Nicer startup banner
- Error messages when proxy port is already in use
- `costwatch report --last 24h` formatting

---

## How to Test Right Now

From this directory:

```bash
# Terminal 1: Start the proxy
./costwatch start

# Terminal 2: Send a test API call through the proxy
curl -X POST http://localhost:5555/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: YOUR_ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "X-CostWatch-Tag: test" \
  -d '{
    "model": "claude-haiku-4-5-20251001",
    "max_tokens": 50,
    "messages": [{"role": "user", "content": "Say hello in 5 words"}]
  }'

# Terminal 2: Check the report
./costwatch report
```

If the API call works, the report should show the cost of that one call.

---

## File Map

| File | What it does |
|------|-------------|
| `cmd/costwatch/main.go` | CLI entrypoint — all commands defined here |
| `internal/proxy/server.go` | HTTP server setup, routing |
| `internal/proxy/handler.go` | Non-streaming Anthropic request handling |
| `internal/proxy/streaming.go` | SSE streaming Anthropic request handling |
| `internal/storage/sqlite.go` | All database operations |
| `internal/storage/schema.go` | SQL table + index definitions |
| `internal/pricing/engine.go` | Cost calculation from model + tokens |
| `internal/pricing/pricing.yaml` | Model pricing data |
| `internal/config/config.go` | Config loading from ~/.costwatch/config.yaml |
| `internal/report/terminal.go` | Terminal table formatting |
| `internal/dashboard/` | Empty — dashboard goes here |
| `sdks/python/` | Python SDK wrapper |
| `sdks/typescript/` | TypeScript SDK wrapper |
