# Tirith — How It Works (Plain English)

A non-technical guide to what happens under the hood, where data lives, when things go stale, and when you need to update or clean up.

---

## The Big Picture

Tirith is a single file you run on your computer. When it's running, it sits between your app and the AI provider (like Anthropic). Your app talks to Tirith, Tirith talks to Anthropic, and on the way back it writes down what happened — which model, how many tokens, how much it cost. Your app never knows Tirith is there.

```
Your App  ──>  Tirith (localhost)  ──>  Anthropic
              writes a log entry
              to a local database
```

There are three pieces:

1. **The proxy** — the middleman that forwards API calls and logs them
2. **The database** — a local file that stores every logged call
3. **The dashboard** — a web page (served locally) that shows charts and tables from the database

---

## What's Baked Into the Binary

When Tirith is compiled (built into a single file you can run), two things get permanently embedded inside it:

### The dashboard UI

The charts, layout, and styling of the web dashboard are frozen at build time. They're packed into the binary and served from memory when you open `localhost:5556`. You can't change the dashboard without rebuilding the whole binary.

**What this means for you:** If a dashboard bug is fixed or a new chart is added, you need to download/install a new version of Tirith to see it.

### The pricing table

Model prices (e.g. "Claude Sonnet costs $3 per million input tokens") are stored in a YAML file that gets baked into the binary at build time. There is no auto-update — the binary uses whatever prices were current when it was compiled.

**What this means for you:** If Anthropic or OpenAI changes their prices, Tirith will calculate costs using the old prices until you update to a newer version. The costs in your reports and dashboard will be wrong (not by much, usually, but wrong).

---

## What Lives on Your Computer at Runtime

### The database (`~/.tirith/data.db`)

A SQLite file in your home directory. Every API call that passes through the proxy gets a row: timestamp, model, tokens in, tokens out, cost, latency, tags, etc.

**Key things to know:**

- **It grows forever.** There's no automatic cleanup or expiry. If you make 1,000 API calls a day, you'll have 365,000 rows after a year. The file will be a few hundred MB at most — SQLite handles this fine, but queries over very long time ranges will get slower.
- **To clear it out:** Delete `~/.tirith/data.db` and restart the proxy. It creates a fresh one automatically.
- **To back it up:** Just copy the file. It's a single, self-contained SQLite database.
- **Schema changes between versions:** If a new version of Tirith adds a column to the database, your existing database won't get that column automatically. There's no migration system yet. If things look off after an upgrade, deleting the database and starting fresh is the safest path (you lose history, though).

### The config file (`~/.tirith/config.yaml`)

Optional. If it doesn't exist, Tirith uses sensible defaults (proxy on port 5555, dashboard on port 5556, database at `~/.tirith/data.db`).

**Key things to know:**

- Config is read once when the proxy starts. If you edit the file while the proxy is running, nothing changes until you restart.
- You'd edit this to change ports (if 5555 is already taken by something else) or to point at a different database location.

---

## How the Proxy Runs

When you run `tirith start`, it starts a foreground process — meaning it runs in your terminal and stays there. Close the terminal (or hit Ctrl+C), and the proxy stops.

There's no background daemon, no PID file, no service manager integration. It's the simplest possible setup: run it in a terminal tab.

**Port conflicts:** If something else is already using port 5555, the proxy fails to start with an "address already in use" error. Fix: close whatever's using that port, or change the port in your config file.

**Graceful shutdown:** When you hit Ctrl+C, the proxy finishes any in-flight requests (up to 5 seconds) before shutting down. Nothing gets corrupted.

---

## How the SDK Wrappers Work

The Python and TypeScript wrappers exist so your app doesn't need to know about the proxy. You add one import line, and the wrapper silently redirects all AI client traffic through the proxy.

### What happens at import time

1. The wrapper checks where the proxy should be: first the `TIRITH_PROXY_URL` environment variable, then `~/.tirith/config.yaml`, then falls back to `http://localhost:5555`.
2. It pings the proxy's `/health` endpoint (1-second timeout).
3. **If the proxy is running:** It monkey-patches the AI SDK constructors so new clients default to routing through the proxy.
4. **If the proxy is NOT running:** It logs a warning and does nothing. Your app connects directly to the provider as normal — no cost tracking, but no breakage either.

### When this can go stale

- **Proxy starts after your app:** If your app imports `tirith` before the proxy is running, the health check fails and patching is skipped. Your app connects directly to the provider for its entire lifetime. Fix: start the proxy first, then your app.
- **Proxy dies while your app is running:** The SDK doesn't re-check. If the proxy crashes mid-session, API calls will start failing (connection refused to localhost). Fix: restart the proxy, or restart your app so it falls back to direct connections.
- **Config changes:** If you change the proxy port in config.yaml, you need to restart both the proxy and your app for the SDK to pick up the new port.

---

## How the Dashboard Gets Its Data

The dashboard is a web page at `localhost:5556`. It polls the backend for fresh data every 10 seconds.

Each refresh fires 6 queries against the SQLite database in parallel:
- Overview stats (total spend, call count)
- Daily spend over time
- Spend per model per day
- Model breakdown
- List of models seen
- Recent API calls

**What this means:**

- The dashboard is always ~10 seconds behind reality. It's not truly real-time, but close enough that it feels live.
- There's no server-side caching — every poll hits the database directly. This is fine for one person viewing the dashboard, but it's worth knowing.
- If the proxy isn't running, the dashboard still works (it reads from the database). It just won't show new data until the proxy is running and logging calls again.

---

## When You Need to Update / Reinstall

| Situation | What to do |
|-----------|------------|
| AI provider changes their prices | Install a newer version of Tirith (pricing is baked into the binary) |
| Dashboard has a bug or you want a new feature | Install a newer version of Tirith (UI is baked into the binary) |
| Database seems corrupted or queries are slow | Delete `~/.tirith/data.db` and restart (you lose history) |
| Changed config but proxy isn't picking it up | Restart the proxy (`Ctrl+C` then `tirith start` again) |
| Port 5555 is in use by something else | Edit `~/.tirith/config.yaml` to set a different port, then restart |
| Proxy started after your app | Restart your app so the SDK wrapper re-checks for the proxy |
| Upgraded Tirith and dashboard looks weird | Hard-refresh your browser (`Cmd+Shift+R`) to clear cached assets |
| Upgraded Tirith and reports show unexpected data | Delete `~/.tirith/data.db` if the schema changed (check release notes) |

---

## What Tirith Does NOT Store

- **Prompts and responses** — never logged. Only metadata (model, token counts, cost, latency, tags).
- **API keys** — passed straight through to the provider, never written to disk.
- **Anything in the cloud** — everything stays on your machine in `~/.tirith/`.
