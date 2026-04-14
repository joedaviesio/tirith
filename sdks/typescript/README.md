# Tirith TypeScript SDK

One import, full AI API cost transparency.

## Install

```bash
npm install tirith-sdk
```

Requires the Tirith proxy running locally — install with `brew install joedaviesio/homebrew-tap/tirith`, then `tirith start`.

## Usage

```ts
import "tirith-sdk"; // auto-patches @anthropic-ai/sdk at import time

import Anthropic from "@anthropic-ai/sdk";
const client = new Anthropic(); // calls are transparently logged
```

No other code changes needed. If the proxy isn't running, the SDK logs a warning and connects directly to Anthropic.

## Supported clients

- `@anthropic-ai/sdk` — ✅
- `openai` — 🚧 planned

## See also

Full docs, CLI usage, and dashboard: https://github.com/joedaviesio/tirith

## License

MIT
