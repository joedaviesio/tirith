/**
 * Configuration for the Tirith SDK wrapper.
 */

import { readFileSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";

const DEFAULT_PROXY_PORT = 5555;

let proxyUrlOverride: string | undefined;

export function getProxyUrl(): string {
  if (proxyUrlOverride) {
    return proxyUrlOverride;
  }

  // Check env var first.
  const envUrl = process.env.COSTWATCH_PROXY_URL;
  if (envUrl) {
    return envUrl;
  }

  // Try to read from config file.
  try {
    const configPath = join(homedir(), ".costwatch", "config.yaml");
    const content = readFileSync(configPath, "utf-8");
    // Simple YAML parsing for the port field — avoids a yaml dependency.
    const portMatch = content.match(/^\s*port:\s*(\d+)/m);
    if (portMatch) {
      return `http://localhost:${portMatch[1]}`;
    }
  } catch {
    // Config file doesn't exist — use default.
  }

  return `http://localhost:${DEFAULT_PROXY_PORT}`;
}

export function setProxyUrl(url: string): void {
  proxyUrlOverride = url;
}

export async function isProxyRunning(proxyUrl: string): Promise<boolean> {
  try {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 1000);
    const resp = await fetch(`${proxyUrl}/health`, { signal: controller.signal });
    clearTimeout(timeout);
    return resp.ok;
  } catch {
    return false;
  }
}
