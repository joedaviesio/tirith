# How Monitoring Works

CostWatch is a **transparent proxy**, not a network sniffer. It only sees API calls that are explicitly routed through it.

## The proxy model

```
Your App  ──▶  CostWatch Proxy (localhost:5555)  ──▶  AI Provider (api.anthropic.com)
          ◀──                                    ◀──
```

CostWatch sits in the middle of the request path. When a request passes through, it:

1. Reads custom `X-CostWatch-*` headers (tag, user, session, environment) and strips them
2. Forwards the request to the real AI provider with auth headers intact
3. Reads the response (or streams SSE chunks in real-time)
4. Extracts token counts and calculates cost from the `usage` field
5. Logs the metadata to SQLite — model, tokens, cost, latency, tags
6. Returns the response to your app unchanged

## What gets monitored

Only requests that pass through the proxy. A request reaches the proxy when:

- The SDK wrapper patches the client's base URL to `localhost:5555`
- `costwatch run` injects `*_BASE_URL` env vars into the subprocess
- You manually set `ANTHROPIC_BASE_URL` (or equivalent) to point at the proxy

If none of these are in place, your app talks directly to the AI provider and CostWatch never sees the traffic.

## What does NOT get monitored

- API calls from apps that haven't imported the SDK wrapper or set the base URL
- Calls made by other developers on separate machines (each runs their own proxy)
- Calls from other processes on the same machine, unless they are also routed through the proxy
- Calls where the user has explicitly set a custom `base_url` on the client — the SDK wrapper uses `setdefault` semantics and won't override it

## Multi-developer scenarios

### Same machine

If two developers share a machine (e.g., a shared dev server), each should run their own proxy on a different port. Their data is isolated in separate SQLite databases at `~/.costwatch/data.db` (one per user home directory).

### Separate machines

Fully independent. Each dev runs their own proxy, stores their own data, and sees only their own usage. There is no cross-machine communication in local mode.

### Shared visibility

Local mode is single-developer by design. For team-wide dashboards and shared cost data, use CostWatch Cloud (when available), which routes traffic through a hosted proxy and aggregates data across the team.

## Privacy

CostWatch logs **metadata only** by default:

- Model name, provider
- Input/output token counts
- Cost (in cents)
- Latency
- Custom tags

Prompt and response content are **never logged** unless explicitly enabled. API keys are forwarded to the provider and never stored.

## Graceful fallback

If the proxy isn't running, the SDK wrappers detect the connection failure, log a warning, and fall back to calling the AI provider directly. Your app continues to work — you just lose observability until the proxy is restarted.
