"""
Tirith load test — staggered API calls over 15 minutes.
Tests multiple models, tags, streaming modes. Hard budget cap at $2.30.
"""

import os
import sys
import time
import json
import random
import urllib.request
from pathlib import Path
from dotenv import load_dotenv
from anthropic import Anthropic

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------
BUDGET_CAP_CENTS = 230          # $2.30 hard cap (leaves $0.20 margin)
TEST_DURATION_MINUTES = 15
PROXY_URL = "http://localhost:5555"
DASHBOARD_URL = "http://localhost:5556"

# Cost per million tokens (from pricing.yaml)
PRICING = {
    "claude-haiku-4-5-20251001": {"input": 0.80, "output": 4.00},
    "claude-sonnet-4-6":        {"input": 3.00, "output": 15.00},
    "claude-opus-4-6":          {"input": 15.00, "output": 75.00},
}

# Test schedule: (model, tag, user, environment, streaming, prompt)
PROMPTS = [
    # --- Haiku batch (cheap, frequent) ---
    ("claude-haiku-4-5-20251001", "health-check",  "bot",     "test", False, "Reply with exactly: OK"),
    ("claude-haiku-4-5-20251001", "summarizer",     "alice",   "test", False, "Summarize the concept of API proxies in 2 sentences."),
    ("claude-haiku-4-5-20251001", "summarizer",     "bob",     "test", True,  "Summarize what a reverse proxy does in 3 bullet points."),
    ("claude-haiku-4-5-20251001", "classifier",     "alice",   "test", False, "Classify this as positive, negative, or neutral: 'The proxy is working great.'"),
    ("claude-haiku-4-5-20251001", "health-check",   "bot",     "test", False, "Reply with exactly: OK"),
    ("claude-haiku-4-5-20251001", "extractor",      "bob",     "staging", False, "Extract the programming languages from: 'We use Go, Python, and TypeScript.'"),
    ("claude-haiku-4-5-20251001", "summarizer",     "charlie", "test", True,  "What is SQLite? Answer in one paragraph."),
    ("claude-haiku-4-5-20251001", "classifier",     "alice",   "test", False, "Is this a question or a statement: 'The dashboard looks clean.'"),
    ("claude-haiku-4-5-20251001", "health-check",   "bot",     "test", False, "Reply with exactly: OK"),
    ("claude-haiku-4-5-20251001", "extractor",      "charlie", "test", False, "List the HTTP methods mentioned: 'We use GET, POST, and DELETE endpoints.'"),
    ("claude-haiku-4-5-20251001", "summarizer",     "alice",   "staging", True,  "Explain cost observability in 2 sentences."),
    ("claude-haiku-4-5-20251001", "health-check",   "bot",     "test", False, "Reply with exactly: OK"),
    ("claude-haiku-4-5-20251001", "classifier",     "bob",     "test", False, "Classify the sentiment: 'This test is taking forever.'"),
    ("claude-haiku-4-5-20251001", "extractor",      "alice",   "test", False, "Extract numbers from: 'The proxy handles 5555 on port and dashboard on 5556.'"),
    ("claude-haiku-4-5-20251001", "summarizer",     "charlie", "test", False, "What is SSE streaming? One sentence."),
    ("claude-haiku-4-5-20251001", "health-check",   "bot",     "test", False, "Reply with exactly: OK"),
    ("claude-haiku-4-5-20251001", "classifier",     "bob",     "staging", True,  "Is this technical or non-technical: 'The WAL mode improves concurrent reads.'"),
    ("claude-haiku-4-5-20251001", "summarizer",     "alice",   "test", False, "Define 'monkey-patching' in one sentence."),
    ("claude-haiku-4-5-20251001", "health-check",   "bot",     "test", False, "Reply with exactly: OK"),
    ("claude-haiku-4-5-20251001", "extractor",      "charlie", "test", False, "Extract the file extensions: 'We write .go, .py, .ts, and .tsx files.'"),

    # --- Sonnet batch (mid-tier, moderate) ---
    ("claude-sonnet-4-6", "analysis",    "alice",   "test",    False, "Compare REST APIs vs GraphQL in 3 bullet points. Be concise."),
    ("claude-sonnet-4-6", "analysis",    "bob",     "test",    True,  "What are the top 3 causes of memory leaks in Go programs? Brief answer."),
    ("claude-sonnet-4-6", "generator",   "charlie", "staging", False, "Write a 4-line Python function that checks if a port is in use."),
    ("claude-sonnet-4-6", "analysis",    "alice",   "test",    False, "What is connection pooling and why does it matter? 2 sentences."),
    ("claude-sonnet-4-6", "generator",   "bob",     "test",    True,  "Write a single Go function that returns the current memory usage in MB."),
    ("claude-sonnet-4-6", "analysis",    "charlie", "test",    False, "Explain the difference between SQLite WAL and journal mode in 2 sentences."),
    ("claude-sonnet-4-6", "generator",   "alice",   "staging", False, "Write a 3-line bash one-liner that checks if localhost:5555 is responding."),
    ("claude-sonnet-4-6", "analysis",    "bob",     "test",    True,  "What is an AbortController in JavaScript? One paragraph."),
    ("claude-sonnet-4-6", "generator",   "charlie", "test",    False, "Write a TypeScript type for an API call log entry with: id, model, cost, timestamp."),
    ("claude-sonnet-4-6", "analysis",    "alice",   "test",    False, "Why do we use embed.FS in Go? 2 sentences."),

    # --- Opus batch (expensive, few) ---
    ("claude-opus-4-6", "deep-analysis", "alice",   "test",    True,  "Describe 3 architectural patterns for building a CLI tool that also serves a web dashboard. Be concise — 1 paragraph each."),
    ("claude-opus-4-6", "deep-analysis", "bob",     "test",    False, "What are the security implications of a transparent API proxy? List 4 key concerns in 1 sentence each."),
    ("claude-opus-4-6", "deep-analysis", "charlie", "staging", True,  "Compare SQLite vs PostgreSQL for a local-first developer tool. 3 bullet points."),
    ("claude-opus-4-6", "deep-analysis", "alice",   "test",    False, "Explain how SDK monkey-patching works for API observability. Keep it to 1 paragraph."),
    ("claude-opus-4-6", "deep-analysis", "bob",     "test",    True,  "What would a pricing engine need to handle for multi-provider AI cost tracking? 4 bullet points."),
]

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def estimate_cost_cents(model: str, input_tokens: int, output_tokens: int) -> float:
    p = PRICING[model]
    return (input_tokens * p["input"] / 1_000_000 + output_tokens * p["output"] / 1_000_000) * 100


def get_tirith_cost() -> int:
    """Fetch current total cost from Tirith dashboard API."""
    try:
        req = urllib.request.Request(f"{DASHBOARD_URL}/api/overview?last=1h")
        with urllib.request.urlopen(req, timeout=3) as resp:
            data = json.loads(resp.read())
            return data.get("total_cost_cents", 0)
    except Exception:
        return -1


def print_status(i: int, total: int, model: str, tag: str, streaming: bool,
                 cost_cents: float, cumulative_cents: float, elapsed: float):
    mode = "stream" if streaming else "sync"
    bar_len = 30
    filled = int(bar_len * (i + 1) / total)
    bar = "█" * filled + "░" * (bar_len - filled)
    pct = (i + 1) / total * 100

    print(f"\n  [{bar}] {pct:5.1f}%  ({i+1}/{total})")
    print(f"  Model:       {model}")
    print(f"  Tag:         {tag:<16} Mode: {mode}")
    print(f"  Est. cost:   ${cost_cents/100:.4f}     Cumulative: ${cumulative_cents/100:.4f}")
    print(f"  Elapsed:     {elapsed:.0f}s / {TEST_DURATION_MINUTES * 60}s")
    print(f"  Budget left: ${(BUDGET_CAP_CENTS - cumulative_cents)/100:.2f} / ${BUDGET_CAP_CENTS/100:.2f}")


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    load_dotenv(Path(__file__).parent / ".env")

    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key or api_key == "your-key-here":
        print("ERROR: Set your ANTHROPIC_API_KEY in testing/.env")
        sys.exit(1)

    # Verify proxy is running
    try:
        req = urllib.request.Request(f"{PROXY_URL}/health")
        with urllib.request.urlopen(req, timeout=2) as resp:
            if resp.status != 200:
                raise Exception()
    except Exception:
        print(f"ERROR: Tirith proxy not running at {PROXY_URL}")
        print("  Run: ./tirith start")
        sys.exit(1)

    client = Anthropic(
        api_key=api_key,
        base_url=f"{PROXY_URL}/proxy/anthropic",
    )

    total = len(PROMPTS)
    interval = (TEST_DURATION_MINUTES * 60) / total  # seconds between calls
    cumulative_cents = 0.0
    start_time = time.time()
    results = []

    print(f"""
  ┌─────────────────────────────────────────────┐
  │  Tirith Load Test                           │
  │                                             │
  │  Calls:    {total:<5}                            │
  │  Duration: {TEST_DURATION_MINUTES} minutes (1 call every ~{interval:.0f}s)     │
  │  Budget:   ${BUDGET_CAP_CENTS/100:.2f}                            │
  │  Models:   Haiku / Sonnet / Opus            │
  │  Dashboard: {DASHBOARD_URL:<24}     │
  └─────────────────────────────────────────────┘
""")

    for i, (model, tag, user, env, streaming, prompt) in enumerate(PROMPTS):
        # Budget check
        if cumulative_cents >= BUDGET_CAP_CENTS:
            print(f"\n  BUDGET CAP REACHED (${cumulative_cents/100:.2f}). Stopping.")
            break

        elapsed = time.time() - start_time

        # Stagger: wait until it's time for this call
        target_time = i * interval
        if elapsed < target_time:
            wait = target_time - elapsed
            time.sleep(wait)

        call_start = time.time()

        try:
            extra_headers = {
                "X-CostWatch-Tag": tag,
                "X-CostWatch-User": user,
                "X-CostWatch-Environment": env,
            }

            if streaming:
                input_tokens = 0
                output_tokens = 0
                with client.messages.stream(
                    model=model,
                    max_tokens=256,
                    messages=[{"role": "user", "content": prompt}],
                    extra_headers=extra_headers,
                ) as stream:
                    response = stream.get_final_message()
                    input_tokens = response.usage.input_tokens
                    output_tokens = response.usage.output_tokens
            else:
                response = client.messages.create(
                    model=model,
                    max_tokens=256,
                    messages=[{"role": "user", "content": prompt}],
                    extra_headers=extra_headers,
                )
                input_tokens = response.usage.input_tokens
                output_tokens = response.usage.output_tokens

            cost = estimate_cost_cents(model, input_tokens, output_tokens)
            cumulative_cents += cost
            latency = (time.time() - call_start) * 1000

            results.append({
                "model": model, "tag": tag, "user": user,
                "input": input_tokens, "output": output_tokens,
                "cost_cents": round(cost, 4), "latency_ms": round(latency),
                "streaming": streaming, "status": "ok",
            })

            print_status(i, total, model, tag, streaming, cost, cumulative_cents, time.time() - start_time)

        except Exception as e:
            print(f"\n  ERROR on call {i+1}: {e}")
            results.append({
                "model": model, "tag": tag, "status": "error", "error": str(e),
            })

    # ---------------------------------------------------------------------------
    # Summary
    # ---------------------------------------------------------------------------
    elapsed_total = time.time() - start_time
    tirith_cost = get_tirith_cost()

    ok = [r for r in results if r["status"] == "ok"]
    errors = [r for r in results if r["status"] == "error"]

    print(f"""
  ┌─────────────────────────────────────────────┐
  │  Test Complete                               │
  ├─────────────────────────────────────────────┤
  │  Calls made:      {len(results):<5}                      │
  │  Successful:      {len(ok):<5}                      │
  │  Errors:          {len(errors):<5}                      │
  │  Duration:        {elapsed_total/60:.1f} min                    │
  │  Est. cost:       ${cumulative_cents/100:.4f}                   │
  │  Tirith tracked:  ${tirith_cost/100:.4f} {'(from API)' if tirith_cost >= 0 else '(unavailable)'}          │
  └─────────────────────────────────────────────┘
""")

    # Breakdown by model
    print("  Cost by model:")
    for model in PRICING:
        model_calls = [r for r in ok if r["model"] == model]
        if model_calls:
            model_cost = sum(r["cost_cents"] for r in model_calls)
            avg_latency = sum(r["latency_ms"] for r in model_calls) / len(model_calls)
            short = model.split("-")[1] if "-" in model else model
            print(f"    {model:<35} {len(model_calls):>3} calls  ${model_cost/100:.4f}  avg {avg_latency:.0f}ms")

    print("\n  Cost by tag:")
    tags = sorted(set(r["tag"] for r in ok))
    for t in tags:
        tag_calls = [r for r in ok if r["tag"] == t]
        tag_cost = sum(r["cost_cents"] for r in tag_calls)
        print(f"    {t:<20} {len(tag_calls):>3} calls  ${tag_cost/100:.4f}")

    print(f"\n  Dashboard: {DASHBOARD_URL}")
    print()


if __name__ == "__main__":
    main()
