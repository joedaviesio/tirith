# Tirith — CLAUDE.md

## Project Overview

Tirith is an open-source CLI + transparent proxy for AI API cost observability. It sits between apps and AI providers (Anthropic, OpenAI, Google), logging every call with cost, tokens, latency, and custom tags. SDK wrappers auto-patch clients at import time — zero config, no env var changes.

Two modes: **Local CLI** (free, SQLite, localhost dashboard) and **Cloud SaaS** (hosted proxy, team dashboards, alerts).

See `TIRITH_ARCHITECTURE.md` for full architecture, schema, and product details.

## Tech Stack

- **CLI + Proxy:** Go (single binary, cross-platform)
- **CLI framework:** Cobra (`github.com/spf13/cobra`)
- **Local storage:** SQLite (pure Go via `modernc.org/sqlite`)
- **Proxy:** `net/http/httputil.ReverseProxy`
- **Dashboard:** React app embedded into Go binary via `embed`
- **Terminal output:** `github.com/olekukonez/tablewriter`
- **Python SDK wrapper:** pip package, monkey-patches Anthropic/OpenAI clients
- **TypeScript SDK wrapper:** npm package, monkey-patches @anthropic-ai/sdk and openai

## Project Structure

```
tirith/
├── cmd/tirith/main.go             # CLI entrypoint (Cobra)
├── internal/
│   ├── proxy/                     # HTTP proxy server + SSE streaming
│   ├── pricing/                   # Cost calculation engine
│   ├── storage/                   # SQLite operations + schema
│   ├── dashboard/                 # Dashboard HTTP server + embedded React
│   └── report/                    # Terminal output formatting
├── sdks/
│   ├── python/                    # Python SDK wrapper (pip install tirith)
│   │   ├── pyproject.toml
│   │   └── tirith/                # __init__.py auto-patches on import
│   └── typescript/                # TypeScript SDK wrapper (npm install tirith)
│       ├── package.json
│       └── src/                   # index.ts auto-patches on import
├── pricing/pricing.yaml           # Community-maintained model pricing
├── TIRITH_ARCHITECTURE.md         # Full architecture reference
├── go.mod / go.sum
└── Makefile
```

## Development Commands

```bash
# Go CLI
go build -o tirith ./cmd/tirith
go run ./cmd/tirith start
go run ./cmd/tirith report
go run ./cmd/tirith run -- python app.py
go test ./...
golangci-lint run

# Python SDK
cd sdks/python && pip install -e .
cd sdks/python && pytest

# TypeScript SDK
cd sdks/typescript && npm install && npm run build
cd sdks/typescript && npm test
```

## Key Design Decisions

- **Zero-config integration:** SDK wrappers monkey-patch official clients at import time. Users add one import line — nothing else changes.
- **Graceful fallback:** If proxy isn't running, SDK wrappers log a warning and connect directly to the provider.
- **`setdefault` semantics:** SDK patches never override an explicit `base_url` set by the user.
- **Privacy-first:** Never log prompt/response content by default. Metadata only.
- **API keys pass through:** Proxy forwards auth headers without storing them.
- **Cost in cents:** Store `cost_cents` as integer to avoid float precision issues.
- **Streaming-first:** SSE chunks forwarded in real-time, token counts accumulated from final usage event.
- **Custom headers stripped:** `X-Tirith-*` headers are read then removed before forwarding to providers.

## Three Integration Paths

1. **SDK wrapper (recommended):** `import tirith` — auto-patches all AI clients
2. **CLI wrapper:** `tirith run -- python app.py` — injects env vars into subprocess
3. **Manual env var:** `export ANTHROPIC_BASE_URL=http://localhost:5555/proxy/anthropic`

## Proxy Routing

Currently only Anthropic is implemented in the proxy:

```
localhost:5555/proxy/anthropic/v1/messages  →  api.anthropic.com/v1/messages
localhost:5555/v1/messages                  →  api.anthropic.com/v1/messages  (auto-detect)
localhost:5555/health                       →  health check
```

OpenAI/Google proxy routes are not yet implemented (SDK wrappers patch OpenAI clients but the proxy only forwards to Anthropic).

## Custom Headers

```
X-Tirith-Tag          # feature tag (e.g., "grant-scanner")
X-Tirith-User         # user attribution
X-Tirith-Session      # session grouping
X-Tirith-Environment  # production, staging, dev
```

## MVP Scope (v0.1)

Focus: Anthropic Messages API only. Go CLI with `start`, `stop`, `report`, `dashboard`, `run`. HTTP proxy with SSE streaming passthrough. Token + cost logging to SQLite. Tag support. Terminal report. Local dashboard. Python + TypeScript SDK wrappers.

Out of scope for MVP: OpenAI/Google proxy routes, `tirith.tag()` context manager, `wrap()` explicit mode, cloud proxy, auth/billing, alerts.

Note: Python SDK already patches OpenAI clients (sync + async), but the Go proxy only routes Anthropic traffic.

## Implementation Status

1. ~~Cobra CLI with `start`, `report`, and `run` commands~~ — done
2. ~~Reverse proxy for Anthropic Messages API (non-streaming)~~ — done
3. ~~Parse response `usage` field for token counts~~ — done
4. ~~Log to SQLite~~ — done
5. ~~`report` command with terminal table~~ — done
6. ~~Streaming (SSE) support~~ — done
7. ~~Python SDK wrapper (`import tirith` auto-patches Anthropic + OpenAI)~~ — done
8. ~~TypeScript SDK wrapper (`import "tirith"` auto-patches @anthropic-ai/sdk)~~ — done
9. ~~`tirith run -- <cmd>` CLI wrapper~~ — done
10. ~~Build and embed dashboard~~ — done (Next.js + Recharts, embedded via `go:embed`)
11. ~~Tag support via custom headers~~ — done
12. Package for distribution (Go binary via brew, Python via PyPI, TS via npm) — in progress

## Dashboard API Endpoints

The embedded dashboard server (port 5556) exposes:

```
GET /api/overview        — total spend, call count, date range
GET /api/by-model        — cost breakdown by model
GET /api/by-tag          — cost breakdown by tag
GET /api/by-user         — cost breakdown by user
GET /api/daily-spend     — daily spend time series
GET /api/daily-by-model  — daily spend broken out by model
GET /api/calls           — paginated list of recent calls
GET /api/models          — list of models seen
GET /api/proxy-health    — proxy liveness check
```

## Ports

- Proxy: `5555` (default, configurable via `--port`)
- Dashboard: `5556` (default)
- Config: `~/.tirith/config.yaml`
- Data: `~/.tirith/data.db`

## Conventions

- Use Go standard library where possible; minimize dependencies.
- Error handling: return errors, don't panic. Wrap errors with context.
- Logging: structured logging (slog or zerolog).
- Tests: table-driven tests, test proxy with httptest.
- No CGO — use `modernc.org/sqlite` for pure Go SQLite (simpler cross-compilation).
- SDK wrappers are separate packages in `sdks/` — they don't depend on the Go binary at build time, only at runtime.
