# Tirith

[![Go Reference](https://pkg.go.dev/badge/github.com/joedaviesio/tirith.svg)](https://pkg.go.dev/github.com/joedaviesio/tirith)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/joedaviesio/tirith)](https://github.com/joedaviesio/tirith/releases)

**AI API cost observability — one import, full transparency.**

Tirith is a local CLI + transparent proxy that logs every AI API call with cost, tokens, latency, and custom tags. Add one import line to your app — nothing else changes.

## Quickstart (30 seconds)

```bash
# 1. Install the binary
brew install joedaviesio/homebrew-tap/tirith

# 2. Start the proxy + dashboard
tirith start

# 3. Install the SDK and add one line to your code
pip install tirith-sdk         # Python
npm install tirith-sdk         # TypeScript
```

```python
import tirith          # auto-patches Anthropic + OpenAI clients
import anthropic

client = anthropic.Anthropic()
client.messages.create(...)  # logged automatically
```

```bash
# 4. See your costs
tirith report
open http://localhost:5556    # dashboard
```

## Integration Paths

Pick whichever fits your workflow — all three do the same thing:

| Path | How | When to use |
|---|---|---|
| **SDK wrapper** | `import tirith` (Python) / `import "tirith-sdk"` (TS) | Recommended — zero config, auto-patches clients at import |
| **CLI wrapper** | `tirith run -- python app.py` | Ad-hoc runs without touching code |
| **Manual env var** | `export ANTHROPIC_BASE_URL=http://localhost:5555/proxy/anthropic` | Anything that reads `*_BASE_URL` |

## Supported Providers

| Provider | Proxy | Python SDK | TypeScript SDK |
|---|---|---|---|
| Anthropic | ✅ | ✅ | ✅ |
| OpenAI | 🚧 planned | ✅ (patches client) | 🚧 planned |
| Google | 🚧 planned | 🚧 planned | 🚧 planned |

## Tagging Calls

Attribute spend to features, users, or environments via headers:

```
X-Tirith-Tag          # feature tag (e.g., "grant-scanner")
X-Tirith-User         # user attribution
X-Tirith-Session      # session grouping
X-Tirith-Environment  # production, staging, dev
```

## Privacy

- Prompts and responses are **never** logged — metadata only.
- API keys pass through the proxy; Tirith never stores them.
- All data stays local in `~/.tirith/data.db` (SQLite).

## Configuration

- Proxy port: `5555` (configurable via `tirith start --port`)
- Dashboard port: `5556`
- Config file: `~/.tirith/config.yaml`
- Data file: `~/.tirith/data.db`

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for build instructions and PR guidelines. Security issues: see [SECURITY.md](./SECURITY.md).

## License

MIT — see [LICENSE](./LICENSE).
