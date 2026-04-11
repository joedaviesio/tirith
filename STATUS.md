# Tirith — Project Status

## What's Built

### Go CLI + Proxy (compiles and runs)

The core binary is built. You can run it from this directory with `./tirith`.

| Command | Status | What it does |
|---------|--------|-------------|
| `./tirith start` | Working | Starts the proxy on localhost:5555. Accepts API calls, forwards them to Anthropic, logs token counts + cost to SQLite. |
| `./tirith report` | Working | Reads the SQLite database and prints a cost summary table — total spend, breakdown by model, breakdown by tag. Currently shows $0.00 because no calls have been made yet. |
| `./tirith run -- <your command>` | Working | Starts the proxy, sets the right env vars, runs your command, then shuts down the proxy when done. |
| `./tirith stop` | Partial | Tells you how to kill the proxy. No PID file or signal-based stop yet. |

### What the proxy actually does

When your app sends an API call to `localhost:5555` instead of `api.anthropic.com`, the proxy:

1. Reads any `X-Tirith-Tag` / `X-Tirith-User` headers (for labelling)
2. Strips those headers so Anthropic never sees them
3. Forwards the request to Anthropic exactly as-is
4. Reads the response — pulls out token counts from the `usage` field
5. Calculates the dollar cost using embedded pricing data
6. Logs everything to a local SQLite database (`~/.tirith/data.db`)
7. Returns the original response to your app, unmodified

Your app doesn't know the proxy exists. It gets the same response it would from Anthropic directly.

Supports both regular (wait for full response) and streaming (SSE) calls.

### SDK Wrappers

Small packages that make setup even easier — instead of changing env vars, you just add one import line.

- **Python** (`sdks/python/`) — `import tirith` at the top of your app, and all Anthropic/OpenAI client instances automatically route through the proxy. Patches both sync and async clients.
- **TypeScript** (`sdks/typescript/`) — `import "tirith"` does the same for Node.js apps.

Not published to PyPI/npm yet. For now they're installed locally with `pip install -e ./sdks/python`.

### Pricing Engine

Embedded YAML with current pricing for:
- Anthropic: Opus 4.6, Sonnet 4.6, Haiku 4.5
- OpenAI: GPT-4o, GPT-4o-mini, GPT-4.1, o3, o4-mini

Calculates cost in cents (integers, not floats) to avoid rounding issues.

### Local Storage

SQLite database at `~/.tirith/data.db`. Stores every API call with: timestamp, provider, model, token counts (input/output/cache), cost, latency, status code, tags, and environment label.

---

### Dashboard

Local web UI served at `localhost:5556` via `./tirith dashboard`. Built with Next.js + Recharts, embedded into the Go binary via `go:embed`.

**Features:**
- Summary stat cards (total spend, call count, avg latency)
- Daily spend line chart (with per-model breakdown)
- Cost by model and by tag bar charts
- Paginated recent calls table
- Auto-refresh every 10 seconds
- Proxy health check indicator

**API endpoints served by the dashboard:**
- `/api/overview`, `/api/by-model`, `/api/by-tag`, `/api/by-user`
- `/api/daily-spend`, `/api/daily-by-model`
- `/api/calls`, `/api/models`, `/api/proxy-health`

---

## What's NOT Built Yet

### OpenAI/Google proxy routes

The proxy only routes Anthropic traffic. SDK wrappers patch OpenAI clients to point at the proxy, but `/proxy/openai/` is not handled by the Go proxy yet.

### PID-based stop

`./tirith stop` should write/read a PID file so it can actually stop the proxy, instead of telling you to Ctrl+C.

### Global install

Right now the binary only exists in this directory. For it to work as `tirith` from anywhere, it needs to be either:
- Copied to `/usr/local/bin/`
- Installed via `go install`
- Published via Homebrew or npm

### SDK features not yet implemented

- `tirith.tag()` context manager for per-call tagging
- `wrap()` explicit mode (opt-in per client instead of global patching)
- `tirith.configure()` exists in code but is not properly exported in either SDK

---

## Remaining Work

- [ ] OpenAI proxy route in Go (`/proxy/openai/`)
- [ ] PID file for `tirith stop`
- [ ] Export `configure()` from Python and TypeScript SDKs
- [ ] `tirith.tag()` context manager
- [ ] `wrap()` explicit mode for per-client patching
- [ ] Publish to Homebrew, PyPI, npm
- [ ] End-to-end integration tests

---

## How to Test Right Now

From this directory:

```bash
# Terminal 1: Start the proxy
./tirith start

# Terminal 2: Send a test API call through the proxy
curl -X POST http://localhost:5555/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: YOUR_ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "X-Tirith-Tag: test" \
  -d '{
    "model": "claude-haiku-4-5-20251001",
    "max_tokens": 50,
    "messages": [{"role": "user", "content": "Say hello in 5 words"}]
  }'

# Terminal 2: Check the report
./tirith report
```

If the API call works, the report should show the cost of that one call.

---

## File Map

| File | What it does |
|------|-------------|
| `cmd/tirith/main.go` | CLI entrypoint — all commands defined here |
| `internal/proxy/server.go` | HTTP server setup, routing |
| `internal/proxy/handler.go` | Non-streaming Anthropic request handling |
| `internal/proxy/streaming.go` | SSE streaming Anthropic request handling |
| `internal/storage/sqlite.go` | All database operations |
| `internal/storage/schema.go` | SQL table + index definitions |
| `internal/pricing/engine.go` | Cost calculation from model + tokens |
| `internal/pricing/pricing.yaml` | Model pricing data |
| `internal/config/config.go` | Config loading from ~/.tirith/config.yaml |
| `internal/report/terminal.go` | Terminal table formatting |
| `internal/dashboard/server.go` | Dashboard HTTP server + API endpoints |
| `internal/dashboard/frontend/` | Embedded Next.js + Recharts build |
| `sdks/python/` | Python SDK wrapper |
| `sdks/typescript/` | TypeScript SDK wrapper |
