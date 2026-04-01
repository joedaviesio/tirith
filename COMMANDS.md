# Tirith — Command Reference

## CLI Commands

### `tirith start`

Start the proxy server and dashboard.

```bash
tirith start              # default ports: proxy 5555, dashboard 5556
tirith start --port 6000  # custom proxy port
```

The proxy intercepts AI API calls, logs token usage and cost to SQLite, and forwards requests to the upstream provider. The dashboard starts automatically alongside the proxy.

- Proxy: `http://localhost:5555`
- Dashboard: `http://localhost:5556`

### `tirith stop`

Stop a running proxy.

```bash
tirith stop
```

### `tirith report`

Print a cost summary to the terminal.

```bash
tirith report                       # all time
tirith report --last 24h            # last 24 hours
tirith report --last 7d             # last 7 days
tirith report --last 30d            # last 30 days
tirith report --tag grant-scanner   # filter by tag
tirith report --last 7d --tag api   # combine filters
```

### `tirith dashboard`

Open the web dashboard in your default browser.

```bash
tirith dashboard
```

### `tirith run -- <command>`

Start the proxy, inject env vars, run your command, then shut down the proxy when it exits.

```bash
tirith run -- python app.py
tirith run -- node server.js
tirith run -- go run .
```

Sets `ANTHROPIC_BASE_URL` and `OPENAI_BASE_URL` automatically so your app routes through the proxy with zero code changes.

## Proxy Routes

| Route | Upstream |
|---|---|
| `/proxy/anthropic/v1/messages` | `api.anthropic.com/v1/messages` |
| `/v1/messages` (auto-detect) | `api.anthropic.com/v1/messages` |
| `/health` | Health check (returns `ok`) |

## Custom Headers

Tag your API calls for filtering in reports and the dashboard.

```
X-Tirith-Tag: grant-scanner       # feature tag
X-Tirith-User: user_abc123        # user attribution
X-Tirith-Session: sess_xyz        # session grouping
X-Tirith-Environment: production  # environment label
```

These headers are stripped before forwarding to the upstream provider.

## SDK Integration

### Python

```bash
pip install tirith
```

```python
import tirith  # patches Anthropic + OpenAI SDKs automatically

from anthropic import Anthropic
client = Anthropic()  # routes through Tirith proxy
```

### TypeScript

```bash
npm install tirith
```

```typescript
import "tirith";  // patches @anthropic-ai/sdk and openai automatically

import Anthropic from "@anthropic-ai/sdk";
const client = new Anthropic();  // routes through Tirith proxy
```

### Manual (env vars)

```bash
export ANTHROPIC_BASE_URL=http://localhost:5555/proxy/anthropic
export OPENAI_BASE_URL=http://localhost:5555/proxy/openai
```

## Configuration

Config file: `~/.tirith/config.yaml`

```yaml
proxy:
  port: 5555
  log_bodies: false

dashboard:
  port: 5556

storage:
  driver: sqlite
  path: ~/.tirith/data.db

providers:
  anthropic:
    upstream: https://api.anthropic.com
  openai:
    upstream: https://api.openai.com
  google:
    upstream: https://generativelanguage.googleapis.com
```

## Build

```bash
make build        # build frontend + Go binary
make build-go     # Go binary only (skip frontend rebuild)
make test         # run Go tests
make frontend     # rebuild frontend only
make clean        # remove binary
```

The binary is output as `tirith` in the project root.
