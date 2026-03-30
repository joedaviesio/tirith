# CostWatch — AI API Cost Observability

## Vision

An open-source CLI + proxy that gives AI builders real-time, per-feature, per-user cost visibility for their LLM API spend. One import, full cost transparency — no env vars, no base URL changes.

---

## Product Overview

### Problem

AI builders have no granular visibility into API spend. Provider dashboards show aggregate billing only. Teams can't answer: "Which feature costs the most?", "What did user X cost me this month?", or "Would switching models save money?"

### Solution

A transparent proxy that sits between your app and AI providers (Anthropic, OpenAI, Google, etc.), logging every call with cost, token counts, latency, and custom tags. SDK wrappers automatically route traffic through the proxy — zero config, no env var changes.

### Two Modes

1. **Local CLI (open source, free forever)** — developer runs `costwatch start`, proxy runs on localhost, data stored in SQLite, local dashboard served on a port
2. **Cloud Hosted (paid SaaS)** — hosted proxy at `proxy.costwatch.dev`, team dashboards, alerts, multi-environment tracking

---

## Architecture Overview

```
┌─────────────┐       ┌──────────────────┐       ┌─────────────────┐
│  User's App  │──────▶│  CostWatch Proxy │──────▶│  AI Provider    │
│  (any stack) │◀──────│  (local or cloud)│◀──────│  (Anthropic,    │
│              │       │                  │       │   OpenAI, etc.) │
└─────────────┘       └────────┬─────────┘       └─────────────────┘
                               │
                        ┌──────▼──────┐
                        │  Data Store  │
                        │  (SQLite /   │
                        │   Postgres)  │
                        └──────┬──────┘
                               │
                        ┌──────▼──────┐
                        │  Dashboard   │
                        │  (local web  │
                        │   UI / cloud)│
                        └─────────────┘
```

---

## Component Breakdown

### 1. CLI (`costwatch`)

**Language:** Go (single binary, cross-platform, fast startup)

**Commands:**

| Command | Description |
|---------|-------------|
| `costwatch start` | Start the local proxy server |
| `costwatch start --port 5678` | Start on a custom port (default: 5555) |
| `costwatch stop` | Stop the proxy |
| `costwatch dashboard` | Open the local web dashboard in browser |
| `costwatch report` | Print cost summary to terminal |
| `costwatch report --last 7d` | Report for last 7 days |
| `costwatch report --tag grant-scanner` | Report filtered by tag |
| `costwatch config` | View/edit configuration |
| `costwatch cloud login` | Authenticate with cloud service |
| `costwatch cloud sync` | Sync local data to cloud |

**Distribution:**

- `npm install -g costwatch` (wraps Go binary)
- `brew install costwatch`
- GitHub releases (prebuilt binaries for macOS, Linux, Windows)
- `curl -fsSL https://get.costwatch.dev | sh`

**Config file:** `~/.costwatch/config.yaml`

```yaml
proxy:
  port: 5555
  log_bodies: false          # never log prompt/response content by default
  
providers:
  anthropic:
    upstream: https://api.anthropic.com
  openai:
    upstream: https://api.openai.com
  google:
    upstream: https://generativelanguage.googleapis.com

storage:
  driver: sqlite              # sqlite | postgres
  path: ~/.costwatch/data.db

dashboard:
  port: 5556
  
cloud:
  enabled: false
  api_key: ""
  endpoint: https://api.costwatch.dev
```

---

### 2. Proxy Server

**The core engine.** Intercepts, forwards, logs, and returns AI API calls transparently.

**Request Flow:**

```
1. App sends request to localhost:5555/v1/messages
   (with optional X-CostWatch-Tag header)
                    │
2. Proxy reads:     ▼
   - Target provider (from URL path or config)
   - Model name (from request body)
   - Input tokens (estimate from request body)
   - Tags (from X-CostWatch-Tag header)
   - Timestamp
                    │
3. Proxy forwards   ▼
   request to upstream provider
   (strips CostWatch headers, adds auth)
                    │
4. Provider         ▼
   responds
                    │
5. Proxy reads:     ▼
   - Output tokens (from response usage field)
   - Actual input tokens (from response usage field)
   - Model used (may differ if fallback)
   - Latency (time from forward to response)
   - Status code
   - Cache read/write tokens (Anthropic)
                    │
6. Proxy logs       ▼
   full record to data store
                    │
7. Proxy returns    ▼
   original response to app (unmodified)
```

**Proxy Routing:**

```
localhost:5555/proxy/anthropic/v1/messages  →  api.anthropic.com/v1/messages
localhost:5555/proxy/openai/v1/chat/completions  →  api.openai.com/v1/chat/completions
```

Or simpler — auto-detect provider from request shape:
```
localhost:5555/v1/messages  →  Anthropic (messages API shape)
localhost:5555/v1/chat/completions  →  OpenAI (chat completions shape)
```

**Streaming Support:**

Must handle SSE streaming responses. The proxy should:
- Forward SSE chunks to the client in real-time (no buffering the full response)
- Accumulate token counts from the final `message_delta` / `usage` event
- Log the complete record after stream ends

**Custom Headers (how users tag calls):**

```
X-CostWatch-Tag: grant-scanner           # feature tag
X-CostWatch-User: user_abc123            # user attribution
X-CostWatch-Session: sess_xyz            # session grouping
X-CostWatch-Environment: production      # environment label
```

All headers are optional. Proxy strips them before forwarding to provider.

---

### 3. Pricing Engine

**Responsibility:** Convert (model + token counts) → dollar cost.

**Pricing data stored as a versioned JSON/YAML file:**

```yaml
# pricing.yaml — community-maintained
version: "2026-03-29"

providers:
  anthropic:
    models:
      claude-opus-4-6:
        input_per_million: 15.00
        output_per_million: 75.00
        cache_read_per_million: 1.50
        cache_write_per_million: 18.75
      claude-sonnet-4-6:
        input_per_million: 3.00
        output_per_million: 15.00
      claude-haiku-4-5-20251001:
        input_per_million: 0.80
        output_per_million: 4.00

  openai:
    models:
      gpt-4o:
        input_per_million: 2.50
        output_per_million: 10.00
      gpt-4o-mini:
        input_per_million: 0.15
        output_per_million: 0.60
```

**Auto-update:** CLI checks for pricing updates on startup (daily, from GitHub raw file or API endpoint). Community PRs keep it current.

**Cost calculation:**

```
cost = (input_tokens * input_rate / 1_000_000)
     + (output_tokens * output_rate / 1_000_000)
     + (cache_read_tokens * cache_read_rate / 1_000_000)
     + (cache_write_tokens * cache_write_rate / 1_000_000)
```

---

### 4. Data Store

**Local mode:** SQLite at `~/.costwatch/data.db`

**Cloud mode:** PostgreSQL

**Schema:**

```sql
CREATE TABLE api_calls (
    id              TEXT PRIMARY KEY,        -- UUID
    timestamp       DATETIME NOT NULL,
    provider        TEXT NOT NULL,            -- anthropic, openai, google
    model           TEXT NOT NULL,            -- claude-sonnet-4-6, gpt-4o
    
    -- Token counts
    input_tokens    INTEGER NOT NULL,
    output_tokens   INTEGER NOT NULL,
    cache_read_tokens  INTEGER DEFAULT 0,
    cache_write_tokens INTEGER DEFAULT 0,
    
    -- Cost (stored in cents to avoid float issues)
    cost_cents      INTEGER NOT NULL,
    
    -- Performance
    latency_ms      INTEGER NOT NULL,
    status_code     INTEGER NOT NULL,
    streaming       BOOLEAN DEFAULT FALSE,
    
    -- Tags
    tag             TEXT,                     -- feature tag
    user_tag        TEXT,                     -- user attribution
    session_tag     TEXT,                     -- session grouping
    environment     TEXT DEFAULT 'default',   -- production, staging, dev
    
    -- Metadata
    endpoint        TEXT,                     -- /v1/messages, /v1/chat/completions
    
    -- Indexes
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_timestamp ON api_calls(timestamp);
CREATE INDEX idx_tag ON api_calls(tag);
CREATE INDEX idx_user_tag ON api_calls(user_tag);
CREATE INDEX idx_model ON api_calls(model);
CREATE INDEX idx_environment ON api_calls(environment);
CREATE INDEX idx_provider ON api_calls(provider);
```

**Important: Never store prompt or response content by default.** Privacy-first. Optional `--log-bodies` flag for debugging only, with clear warnings.

---

### 5. Dashboard

**Local:** Embedded web UI served by the Go binary at `localhost:5556`. Single-page React app bundled into the binary using `embed`.

**Cloud:** Full web app at `app.costwatch.dev`.

**Dashboard Views:**

#### Overview
- Total spend (today, this week, this month)
- Spend trend chart (line chart, daily granularity)
- Top models by cost (bar chart)
- Top features/tags by cost (bar chart)
- Request volume over time

#### By Feature (Tag)
- Cost per tag over time
- Average cost per call by tag
- Token usage breakdown by tag
- Comparison table: tag, total cost, avg cost/call, call count, avg latency

#### By User
- Cost per user tag
- Highest-cost users
- User cost trends

#### By Model
- Cost comparison across models
- Token efficiency (output tokens per dollar)
- Latency comparison
- "What-if" calculator: "If you switched this tag from Opus to Sonnet, you'd save $X/month"

#### Alerts (Cloud only)
- Daily spend threshold
- Per-tag budget alerts
- Anomaly detection (spike in usage)
- Delivery: Slack webhook, email, or in-dashboard

---

### 6. Cloud Service (Paid Tier)

**Architecture:**

```
┌─────────────┐       ┌──────────────────────┐       ┌─────────────────┐
│  User's App  │──────▶│  Edge Proxy           │──────▶│  AI Provider    │
│  (Railway,   │◀──────│  (Cloudflare Workers  │◀──────│                 │
│   Render)    │       │   or Fly.io edge)     │       └─────────────────┘
└─────────────┘       └────────┬──────────────┘
                               │
                        ┌──────▼──────┐
                        │  API Service │
                        │  (Go, on     │
                        │   Railway)   │
                        └──────┬──────┘
                               │
                    ┌──────────┼──────────┐
                    │          │          │
              ┌─────▼───┐ ┌───▼────┐ ┌───▼────┐
              │Postgres │ │ Redis  │ │ Queue  │
              │(Neon/   │ │(cache, │ │(async  │
              │Supabase)│ │rate    │ │log     │
              └─────────┘ │limits) │ │writes) │
                          └────────┘ └────────┘
```

**Edge proxy** handles the latency concern — deployed globally so the extra hop is <20ms. It does minimal work: tag extraction, forward to provider, stream response back, and async-post the log record to the API service.

**Cloud API endpoints:**

```
POST   /api/v1/log              — ingest a log record (from edge proxy)
GET    /api/v1/reports/overview  — dashboard overview data
GET    /api/v1/reports/by-tag    — breakdown by tag
GET    /api/v1/reports/by-user   — breakdown by user
GET    /api/v1/reports/by-model  — breakdown by model
POST   /api/v1/alerts            — create/update alert rules
GET    /api/v1/export            — CSV/JSON export
```

**Authentication:**

- API key per workspace: `cw_live_abc123`
- Team management: invite by email, role-based (admin, viewer)
- API key passed as `X-CostWatch-API-Key` header or `COSTWATCH_API_KEY` env var

---

## Integration Guide (How Users Set Up)

### Option A: SDK Wrapper (Recommended — Zero Config)

SDK wrappers monkey-patch the official Anthropic/OpenAI clients at import time, automatically routing all traffic through the CostWatch proxy. The developer adds one import — nothing else changes.

#### Python

```bash
pip install costwatch
```

```python
# Add this ONE line at the top of your app (before creating any clients)
import costwatch  # that's it — patches Anthropic + OpenAI SDKs automatically

# Everything below is unchanged — your existing code just works
from anthropic import Anthropic
client = Anthropic()

response = client.messages.create(
    model="claude-sonnet-4-6",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello"}],
)
# → request transparently routes through CostWatch proxy
# → tokens, cost, latency all logged
```

**How it works under the hood:**

```python
# costwatch/__init__.py (simplified)
import anthropic
import openai

_PROXY = "http://localhost:5555"

# Patch Anthropic — override the default base_url at class level
_original_anthropic_init = anthropic.Anthropic.__init__

def _patched_anthropic_init(self, *args, **kwargs):
    kwargs.setdefault("base_url", f"{_PROXY}/proxy/anthropic")
    _original_anthropic_init(self, *args, **kwargs)

anthropic.Anthropic.__init__ = _patched_anthropic_init

# Patch OpenAI — same approach
_original_openai_init = openai.OpenAI.__init__

def _patched_openai_init(self, *args, **kwargs):
    kwargs.setdefault("base_url", f"{_PROXY}/proxy/openai")
    _original_openai_init(self, *args, **kwargs)

openai.OpenAI.__init__ = _patched_openai_init
```

Key details:
- Uses `setdefault` so explicit `base_url=` in user code is never overridden
- Patches both sync and async clients (`Anthropic`, `AsyncAnthropic`, `OpenAI`, `AsyncOpenAI`)
- Auto-detects proxy port from `~/.costwatch/config.yaml` or `COSTWATCH_PROXY_URL` env var
- If the proxy isn't running, falls through gracefully (logs a warning, connects directly to provider)
- No-op if the SDK isn't installed (e.g., if you only use Anthropic, OpenAI patch is skipped)

#### TypeScript / Node.js

```bash
npm install costwatch
```

```typescript
// Add this ONE line at the top of your entrypoint
import "costwatch";  // patches @anthropic-ai/sdk and openai at import time

// Everything below is unchanged
import Anthropic from "@anthropic-ai/sdk";
const client = new Anthropic();
// → routes through CostWatch proxy automatically
```

**How it works under the hood:**

```typescript
// costwatch/index.ts (simplified)
const PROXY = "http://localhost:5555";

try {
  const anthropicMod = await import("@anthropic-ai/sdk");
  const OriginalAnthropic = anthropicMod.default;
  const originalInit = OriginalAnthropic.prototype.constructor;

  anthropicMod.default = class extends OriginalAnthropic {
    constructor(opts: any = {}) {
      opts.baseURL ??= `${PROXY}/proxy/anthropic`;
      super(opts);
    }
  };
} catch {
  // @anthropic-ai/sdk not installed — skip
}
```

#### Explicit Mode (Opt-In Per Client)

For users who want surgical control instead of global patching:

```python
from costwatch import wrap

from anthropic import Anthropic
client = wrap(Anthropic())  # only this client routes through the proxy
```

```typescript
import { wrap } from "costwatch";
import Anthropic from "@anthropic-ai/sdk";

const client = wrap(new Anthropic());  // only this client
```

### Option B: Env Var Override (No SDK Dependency)

For users who can't or don't want to install the SDK wrapper (e.g., non-Python/TS languages, or preference for explicit config):

```bash
# Install CLI
brew install costwatch  # or: npm install -g costwatch

# Start proxy
costwatch start

# Point your SDK at the proxy
export ANTHROPIC_BASE_URL=http://localhost:5555/proxy/anthropic
export OPENAI_BASE_URL=http://localhost:5555/proxy/openai

# Run your app normally — SDKs read these env vars natively
```

### Option C: CLI Wrapper (Automatic Env Injection)

CostWatch can wrap your app's start command, injecting the env vars automatically:

```bash
costwatch run -- python app.py
costwatch run -- node server.js
costwatch run -- go run .
```

This starts the proxy (if not running), sets `ANTHROPIC_BASE_URL` / `OPENAI_BASE_URL` in the subprocess env, and runs your command. No code changes, no SDK wrapper, no manual env vars.

### Deployed App (Cloud — Railway, Render, Fly.io, etc.)

```bash
pip install costwatch  # or npm install costwatch
```

```python
import costwatch
costwatch.configure(
    proxy_url="https://proxy.costwatch.dev",
    api_key="cw_live_abc123",
)
```

Or via env vars (no code changes):

```
COSTWATCH_PROXY_URL=https://proxy.costwatch.dev
COSTWATCH_API_KEY=cw_live_abc123
```

### Tagging Calls

Tags work the same regardless of integration method:

```python
import costwatch

# Global tags (applied to every call automatically)
costwatch.configure(
    default_tags={
        "environment": "production",
    }
)

# Per-call tags via extra headers (works with any integration method)
response = client.messages.create(
    model="claude-sonnet-4-6",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello"}],
    extra_headers={
        "X-CostWatch-Tag": "grant-scanner",
        "X-CostWatch-User": f"user_{user_id}",
    },
)

# Or via the costwatch helper (cleaner API)
with costwatch.tag("grant-scanner", user=user_id):
    response = client.messages.create(
        model="claude-sonnet-4-6",
        max_tokens=1024,
        messages=[{"role": "user", "content": "Hello"}],
    )
```

---

## Tech Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| CLI + Proxy | Go | Single binary, fast, cross-platform, excellent HTTP/streaming |
| Python SDK wrapper | Python (pip package) | Monkey-patches Anthropic/OpenAI clients at import time |
| TypeScript SDK wrapper | TypeScript (npm package) | Monkey-patches @anthropic-ai/sdk and openai at import time |
| Local storage | SQLite | Zero dependency, portable, embedded |
| Local dashboard | React (embedded via Go embed) | Bundled into binary, no separate process |
| Cloud proxy | Cloudflare Workers or Fly.io edge | Global distribution, low latency |
| Cloud API | Go on Railway | Consistent with CLI codebase |
| Cloud DB | PostgreSQL (Neon or Supabase) | Scalable, familiar, good analytics queries |
| Cloud cache | Redis (Upstash) | Rate limiting, real-time counters |
| Cloud dashboard | Next.js | Fast iteration, SSR for SEO/marketing pages |
| Auth | Clerk or custom JWT | Team management, API key generation |
| Payments | Stripe | Usage-based billing |
| Pricing data | YAML in GitHub repo | Community-maintained, versioned, auto-updated |

---

## Monetization

### Open Source (Free Forever)
- Local CLI + proxy
- SQLite storage
- Local dashboard
- Unlimited tracked calls
- All provider support

### Cloud Pro ($29/mo or usage-based)
- Hosted proxy (low-latency edge)
- Team dashboard (up to 5 seats)
- 30-day data retention
- Slack/email alerts
- CSV/JSON export

### Cloud Team ($79/mo or usage-based)
- Everything in Pro
- Unlimited seats
- 90-day retention
- SSO
- Custom alert rules
- "What-if" model switching calculator
- API access to your cost data

### Usage-Based Alternative
- Free up to $100/mo tracked AI spend
- 0.5% of tracked spend above that
- Aligns incentives: tool cost scales with the value it tracks

---

## MVP Scope (v0.1)

### In Scope
- [ ] Go CLI with `start`, `stop`, `report`, `dashboard`, `run` commands
- [ ] HTTP proxy supporting Anthropic Messages API
- [ ] Streaming (SSE) passthrough
- [ ] Token + cost logging to SQLite
- [ ] `X-CostWatch-Tag` header support
- [ ] Pricing engine with Anthropic models
- [ ] Terminal report output (table format)
- [ ] Basic local dashboard (spend over time, by model, by tag)
- [ ] Python SDK wrapper (`pip install costwatch`) — auto-patches Anthropic client
- [ ] TypeScript SDK wrapper (`npm install costwatch`) — auto-patches @anthropic-ai/sdk
- [ ] `costwatch run -- <cmd>` CLI wrapper (env var injection)
- [ ] `brew install` and `npm install -g` (for CLI binary) distribution

### Out of Scope for MVP
- [ ] OpenAI / Google provider support in SDK wrappers (v0.2)
- [ ] `costwatch.tag()` context manager / `costwatch.configure()` (v0.2)
- [ ] Cloud hosted proxy (v0.3)
- [ ] Team features, auth, billing (v0.4)
- [ ] Alerts and anomaly detection (v0.5)
- [ ] "What-if" model calculator (v0.5)
- [ ] Self-hosted cloud option via Docker (v0.6)

---

## Security & Privacy Principles

1. **Never log prompt/response content by default** — metadata only
2. **API keys pass through, never stored** — proxy forwards auth headers without persisting
3. **Local-first** — all data stays on developer's machine unless they opt into cloud
4. **Open source core** — anyone can audit the proxy code
5. **Minimal permissions** — proxy only needs outbound HTTPS to providers
6. **Cloud data encrypted at rest and in transit**
7. **SOC 2 compliance** target for cloud tier (post-revenue)

---

## Competitive Landscape

| Tool | Approach | Gap |
|------|----------|-----|
| Helicone | Proxy-based, cloud-first | No local/CLI option, heavier setup |
| LiteLLM | Python proxy + model gateway | More about model switching than cost observability |
| Portkey | Enterprise AI gateway | Overkill for indie/small teams |
| LangSmith | Tracing + eval focused | Cost is secondary, tightly coupled to LangChain |
| Provider dashboards | Aggregate billing | No per-feature, per-user, or per-tag granularity |

**CostWatch differentiator:** Local-first, open-source, single-binary CLI. Designed for indie builders and small teams, not enterprises. True zero-config: `import costwatch` and you're done — no env vars, no base URL changes, no code rewrites. Privacy-first (no content logging).

---

## Open Questions

- Go vs Rust for the CLI? Go is faster to build, Rust is more "cool factor" for dev tools
- Should the proxy also support function/tool call cost tracking separately from message tokens?
- Is there demand for a VS Code extension that shows cost inline during development?
- Should the dashboard include prompt/response viewer (opt-in) for debugging, or stay purely cost-focused?
- Naming: CostWatch, TokenMeter, BurnRate, SpendLens — what resonates?
- SDK wrappers: should the Python/TS packages also bundle a lightweight proxy, or require the Go binary to be running separately? (Leaning: require Go binary — keeps SDK packages tiny and avoids duplicating proxy logic)
- Should `costwatch run` auto-install the proxy if not present, or just error with install instructions?
- Graceful fallback: if proxy is down, should SDK wrappers silently connect directly to the provider (current plan), or should there be a strict mode that errors?

---

## Getting Started (for Claude Code)

```bash
# Bootstrap the project
mkdir costwatch && cd costwatch
go mod init github.com/[username]/costwatch

# Suggested directory structure
costwatch/
├── cmd/
│   └── costwatch/
│       └── main.go              # CLI entrypoint (cobra)
├── internal/
│   ├── proxy/
│   │   ├── server.go            # HTTP proxy server
│   │   ├── handler.go           # Request/response interception
│   │   └── streaming.go         # SSE streaming handler
│   ├── pricing/
│   │   ├── engine.go            # Cost calculation
│   │   └── models.yaml          # Pricing data
│   ├── storage/
│   │   ├── sqlite.go            # SQLite operations
│   │   └── schema.go            # Table definitions
│   ├── dashboard/
│   │   ├── server.go            # Dashboard HTTP server
│   │   └── frontend/            # Embedded React app
│   └── report/
│       └── terminal.go          # Terminal output formatting
├── sdks/
│   ├── python/                  # Python SDK wrapper (pip install costwatch)
│   │   ├── pyproject.toml
│   │   ├── costwatch/
│   │   │   ├── __init__.py      # Auto-patches Anthropic/OpenAI on import
│   │   │   ├── _patch.py        # Monkey-patching logic
│   │   │   ├── _config.py       # configure(), proxy discovery
│   │   │   └── _tags.py         # tag() context manager, default_tags
│   │   └── tests/
│   └── typescript/              # TypeScript SDK wrapper (npm install costwatch)
│       ├── package.json
│       ├── src/
│       │   ├── index.ts         # Auto-patches SDKs on import
│       │   ├── patch.ts         # Monkey-patching logic
│       │   └── config.ts        # configure(), proxy discovery
│       └── tests/
├── pricing/
│   └── pricing.yaml             # Community-maintained pricing
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Key Go Libraries
- `github.com/spf13/cobra` — CLI framework
- `github.com/mattn/go-sqlite3` — SQLite driver
- `net/http/httputil` — ReverseProxy for proxying
- `github.com/olekukonez/tablewriter` — Terminal table output
- `embed` — Bundle dashboard frontend into binary

### First Implementation Steps
1. Set up Cobra CLI with `start`, `report`, and `run` commands
2. Build the reverse proxy for Anthropic Messages API (non-streaming first)
3. Parse response `usage` field for token counts
4. Log to SQLite
5. Add `report` command with terminal table output
6. Add streaming support
7. Build Python SDK wrapper (`import costwatch` auto-patches Anthropic)
8. Build TypeScript SDK wrapper (`import "costwatch"` auto-patches @anthropic-ai/sdk)
9. Build `costwatch run -- <cmd>` CLI wrapper
10. Build and embed the dashboard
11. Add tag support via custom headers
12. Package for distribution (Go binary via brew/npm, Python via PyPI, TS via npm)
