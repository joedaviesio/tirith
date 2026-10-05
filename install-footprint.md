# Tirith: Install Footprint & Data Flow

A practical guide to where Tirith puts things on your machine, what's global vs. per-project, and how to remove it cleanly.

## Where Tirith lives on your machine

**Global (user-level) — nothing per-project:**

- `~/.tirith/config.yaml` — optional config (ports, etc.)
- `~/.tirith/data.db` — SQLite database with *every* call logged, across every project
- The `tirith` binary itself — wherever you installed it (`/usr/local/bin/tirith`, `~/go/bin/tirith`, or a Homebrew path)

**Per-project — only if the SDK is installed:**

- `tirith-sdk` in your `requirements.txt` / `package.json` (just a dependency import; no config files written into the project)

That's it. There's no `.tirith/` folder inside your project, no shell rc modifications, no global env var exports.

## Why the scope confusion is real

Tirith **is** machine-wide by design, and here's the subtle part:

- The **proxy** (`tirith start`) runs once for the whole machine on `localhost:5555`. Every project that imports the SDK routes through that single proxy and writes to the *same* `~/.tirith/data.db`.
- So when you run `tirith report` in project A, you'll see calls from project B too — unless you filter with `--tag` or `--environment`.
- The SDK itself doesn't "cover all API keys on your computer" — it only intercepts calls from processes where `import tirith` (Python) or `import "tirith"` (TS) actually runs. Any app that doesn't import the SDK is invisible to Tirith.

Mental model: **one proxy + one database per user**, but **only instrumented apps feed into it**. Tagging (`X-Tirith-Tag: project-a`) is how you slice by project.

## Data flow

```
your app (imports tirith-sdk)
   │
   ▼  patches Anthropic/OpenAI client's base_url → localhost:5555
Tirith proxy (~/tirith binary, running in background)
   │
   ├─▶ forwards request to api.anthropic.com / api.openai.com
   │   (API key passes through; never stored)
   │
   └─▶ logs metadata (model, tokens, cost_cents, latency, tag) to ~/.tirith/data.db
           │
           ▼
       dashboard (localhost:5556) + `tirith report` read from this DB
```

No prompt/response content is logged by default. API keys are forwarded but never written to disk.

## Uninstalling — full cleanup

1. **Stop the proxy:** `tirith stop` (or kill the process)
2. **Remove the binary:**
   - Homebrew: `brew uninstall tirith`
   - Manual: `rm $(which tirith)`
   - Go install: `rm ~/go/bin/tirith`
3. **Remove all data + config:** `rm -rf ~/.tirith`
4. **Uninstall SDKs from each project:**
   - `pip3 uninstall tirith-sdk`
   - `npm uninstall tirith-sdk`
5. **Remove the import lines** from your source files (one `import tirith` / `import "tirith"` per entrypoint)

After that there's zero trace. No system services, no PATH edits, no shell rc entries, no registry keys (Windows).
