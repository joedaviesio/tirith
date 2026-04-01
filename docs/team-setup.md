# Team Setup Guide

How to get each developer on your team running Tirith locally.

## Prerequisites

- Python 3.8+ and/or Node.js 18+ (for SDK wrappers)

## 1. Install the CLI

Each developer installs Tirith on their own machine:

```bash
brew install joedaviesio/tap/tirith
```

## 2. Start Tirith

```bash
tirith
```

That's it. This starts the proxy on `localhost:5555`, the dashboard on `localhost:5556`, and opens the dashboard in your browser automatically.

To start without opening the browser:

```bash
tirith --no-open
```

To use a custom port:

```bash
tirith start --port 5678
```

Stop with `Ctrl+C`.

## 3. Connect your app

Pick one of the three integration paths:

### Option A: SDK wrapper (recommended)

Install the wrapper and add a single import — all AI client calls are automatically routed through the proxy.

**Python:**

```bash
pip install costwatch
```

```python
import costwatch  # add this line — that's it
import anthropic

client = anthropic.Anthropic()
# calls now go through the local proxy
```

**TypeScript:**

```bash
npm install costwatch
```

```typescript
import "costwatch"; // add this line — that's it
import Anthropic from "@anthropic-ai/sdk";

const client = new Anthropic();
// calls now go through the local proxy
```

### Option B: CLI wrapper

Wrap your run command — Tirith injects the right env vars automatically:

```bash
tirith run -- python app.py
tirith run -- node server.js
```

### Option C: Manual env vars

Set the base URL yourself:

```bash
export ANTHROPIC_BASE_URL=http://localhost:5555/proxy/anthropic
python app.py
```

## 4. View your data

The dashboard opens automatically when you run `tirith`. You can also view data in the terminal:

```bash
tirith report
tirith report --last 7d
tirith report --tag grant-scanner
```

## 5. Per-developer data

Each developer gets their own SQLite database at `~/.costwatch/data.db`. There is no shared state between developers in local mode — everyone sees only their own usage.

## Adding tags

Use custom headers to tag requests by feature, user, or environment:

```python
client.messages.create(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello"}],
    extra_headers={
        "X-CostWatch-Tag": "grant-scanner",
        "X-CostWatch-User": "joe",
        "X-CostWatch-Environment": "dev",
    },
)
```
