# Team Setup Guide

How to get each developer on your team running CostWatch locally.

## Prerequisites

- Go 1.21+ (to build from source) or a prebuilt binary
- Python 3.8+ and/or Node.js 18+ (for SDK wrappers)

## 1. Install the CLI

Each developer installs CostWatch on their own machine:

```bash
# From source
git clone <repo-url> && cd costwatch
go build -o costwatch ./cmd/costwatch
mv costwatch /usr/local/bin/

# Or via Homebrew (when published)
brew install costwatch
```

## 2. Start the proxy

```bash
costwatch start
```

This starts the proxy on `localhost:5555` by default. Each developer runs their own independent instance.

To use a custom port:

```bash
costwatch start --port 5678
```

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

Wrap your run command — CostWatch injects the right env vars automatically:

```bash
costwatch run -- python app.py
costwatch run -- node server.js
```

### Option C: Manual env vars

Set the base URL yourself:

```bash
export ANTHROPIC_BASE_URL=http://localhost:5555/proxy/anthropic
python app.py
```

## 4. View your data

```bash
# Terminal report
costwatch report
costwatch report --last 7d
costwatch report --tag grant-scanner

# Local dashboard
costwatch dashboard
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
