/**
 * Tirith — AI API cost observability.
 *
 * Import this module to automatically route all Anthropic and OpenAI API calls
 * through the Tirith proxy for cost tracking.
 *
 * Usage:
 *   import "tirith";  // that's it — patches SDKs automatically
 *
 *   import Anthropic from "@anthropic-ai/sdk";
 *   const client = new Anthropic();  // routes through Tirith proxy
 */

import { getProxyUrl, isProxyRunning } from "./config.js";
import { patchAnthropic, patchOpenAI } from "./patch.js";

let patched = false;

async function patchAll(): Promise<void> {
  const proxyUrl = getProxyUrl();

  const running = await isProxyRunning(proxyUrl);
  if (!running) {
    if (!patched) {
      console.warn(
        `[tirith] Proxy not running at ${proxyUrl} — API calls will go directly to providers. Run 'tirith start' to enable cost tracking.`
      );
    }
    return;
  }

  await patchAnthropic(proxyUrl);
  await patchOpenAI(proxyUrl);
  patched = true;
}

// Auto-patch on import.
patchAll().catch((err) => {
  console.warn(`[tirith] Failed to patch SDKs: ${err}`);
});

export { getProxyUrl } from "./config.js";
