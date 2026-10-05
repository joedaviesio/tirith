# Tirith

Video launch 

https://www.youtube.com/shorts/uZCAqzqLfi4

[![Go Reference](https://pkg.go.dev/badge/github.com/joedaviesio/tirith.svg)](https://pkg.go.dev/github.com/joedaviesio/tirith)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/joedaviesio/tirith)](https://github.com/joedaviesio/tirith/releases)

**AI API cost observability — one import, full transparency.**

Tirith is a local CLI + transparent proxy that logs every AI API call with cost, tokens, latency, and custom tags. Add one import line to your app — nothing else changes.

## Quickstart (30 seconds)

Install the binary, then pick the column that matches your stack:

*shell · `terminal`*

```bash
brew install joedaviesio/homebrew-tap/tirith
```

<table>
<tr>
<th>Anthropic&nbsp;·&nbsp;Proxy</th>
<th>Anthropic&nbsp;·&nbsp;Python&nbsp;SDK</th>
<th>Anthropic&nbsp;·&nbsp;TypeScript&nbsp;SDK</th>
<th>OpenAI&nbsp;·&nbsp;Python&nbsp;SDK</th>
</tr>
<tr>
<td>

*shell · `~/.zshrc`*

```bash
export ANTHROPIC_BASE_URL=\
  http://localhost:5555/proxy/anthropic
```

*shell · `terminal`*

```bash
curl $ANTHROPIC_BASE_URL/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d '{...}'
```

</td>
<td>

*shell · `terminal`*

```bash
pip install tirith-sdk anthropic
```

*python · `app.py`*

```python
import tirith
import anthropic
```

</td>
<td>

*shell · `terminal`*

```bash
npm install tirith-sdk @anthropic-ai/sdk
```

*typescript · `app.ts`*

```ts
import "tirith-sdk";
import Anthropic from "@anthropic-ai/sdk";
```

</td>
<td>

*shell · `terminal`*

```bash
pip install tirith-sdk openai
```

*python · `app.py`*

```python
import tirith
from openai import OpenAI
```

</td>
</tr>
</table>

Once your code is wired up, start the proxy and run your app:

*shell · `terminal`*

```bash
tirith start       # proxy on :5555, dashboard on :5556
```

Then view your spend:

*shell · `terminal`*

```bash
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
