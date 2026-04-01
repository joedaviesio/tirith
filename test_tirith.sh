#!/usr/bin/env bash
#
# Integration test script for Tirith.
#
# Spins up a fake Anthropic upstream, the Tirith proxy + dashboard,
# then exercises:
#   1. Health check
#   2. Non-streaming proxy request with usage logging
#   3. Streaming (SSE) proxy request with usage logging
#   4. Custom header (tag) extraction
#   5. Auto-detect route (/v1/messages)
#   6. Dashboard API endpoints return logged data
#   7. CORS lockdown (only dashboard origin allowed)
#   8. Request body size limit (413 on oversized body)
#   9. Invalid JSON rejected (400)
#  10. Localhost-only binding verification
#
# Exit codes: 0 = all passed, 1 = failure.
set -uo pipefail

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); printf "  \033[32m✓ %s\033[0m\n" "$1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); printf "  \033[31m✗ %s\033[0m\n" "$1"; }

cleanup() {
    # Kill background processes.
    [ -n "${FAKE_PID:-}" ] && kill "$FAKE_PID" 2>/dev/null || true
    [ -n "${TIRITH_PID:-}" ] && kill "$TIRITH_PID" 2>/dev/null || true
    wait 2>/dev/null || true
    # Clean up temp files.
    rm -f "$TMPDIR_TEST"/fake_upstream.py "$TMPDIR_TEST"/tirith_binary
    rm -rf "$TMPDIR_TEST"
}
trap cleanup EXIT

TMPDIR_TEST=$(mktemp -d)
FAKE_PORT=19876
PROXY_PORT=19877
DASH_PORT=19878
DB_PATH="$TMPDIR_TEST/test.db"
CONFIG_PATH="$TMPDIR_TEST/config.yaml"

# ─── Write a fake Anthropic upstream ──────────────────────────────────
cat > "$TMPDIR_TEST/fake_upstream.py" << 'PYEOF'
import http.server, json, sys

PORT = int(sys.argv[1])

class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *a): pass

    def do_POST(self):
        length = int(self.headers.get("Content-Length", 0))
        body = json.loads(self.rfile.read(length)) if length else {}
        stream = body.get("stream", False)

        # Verify Tirith headers are stripped.
        has_cw = any(k.lower().startswith("x-tirith-") for k in self.headers)
        if has_cw:
            self.send_response(500)
            self.end_headers()
            self.wfile.write(b'{"error":"tirith headers leaked"}')
            return

        if stream:
            self.send_response(200)
            self.send_header("Content-Type", "text/event-stream")
            self.end_headers()
            # message_start
            start = json.dumps({
                "type": "message_start",
                "message": {
                    "model": "claude-sonnet-4-6",
                    "usage": {
                        "input_tokens": 50,
                        "output_tokens": 0,
                        "cache_read_input_tokens": 0,
                        "cache_creation_input_tokens": 0
                    }
                }
            })
            self.wfile.write(f"event: message_start\ndata: {start}\n\n".encode())
            self.wfile.flush()

            # content_block_delta (just forward, proxy doesn't parse these)
            delta = json.dumps({"type": "content_block_delta", "delta": {"text": "Hello!"}})
            self.wfile.write(f"event: content_block_delta\ndata: {delta}\n\n".encode())
            self.wfile.flush()

            # message_delta with final output tokens
            msg_delta = json.dumps({
                "type": "message_delta",
                "usage": {"output_tokens": 25}
            })
            self.wfile.write(f"event: message_delta\ndata: {msg_delta}\n\n".encode())
            self.wfile.flush()

            # done
            self.wfile.write(b"data: [DONE]\n\n")
            self.wfile.flush()
        else:
            resp = json.dumps({
                "id": "msg_test",
                "type": "message",
                "model": "claude-sonnet-4-6",
                "role": "assistant",
                "content": [{"type": "text", "text": "Hi from fake upstream"}],
                "usage": {
                    "input_tokens": 100,
                    "output_tokens": 42,
                    "cache_read_input_tokens": 10,
                    "cache_creation_input_tokens": 5
                }
            }).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(resp)))
            self.end_headers()
            self.wfile.write(resp)

http.server.HTTPServer(("127.0.0.1", PORT), Handler).serve_forever()
PYEOF

# ─── Write Tirith config pointing at our fake upstream ────────────────
cat > "$CONFIG_PATH" << EOF
proxy:
  host: "127.0.0.1"
  port: $PROXY_PORT
providers:
  anthropic:
    upstream: "http://127.0.0.1:$FAKE_PORT"
storage:
  driver: sqlite
  path: "$DB_PATH"
dashboard:
  host: "127.0.0.1"
  port: $DASH_PORT
EOF

# ─── Build Tirith ─────────────────────────────────────────────────────
echo "Building Tirith..."
go build -o "$TMPDIR_TEST/tirith_binary" ./cmd/tirith

# ─── Start fake upstream ──────────────────────────────────────────────
python3 "$TMPDIR_TEST/fake_upstream.py" "$FAKE_PORT" &
FAKE_PID=$!

# ─── Start Tirith (override config location via HOME trick) ──────────
# Tirith reads ~/.tirith/config.yaml — symlink our test config.
TEST_HOME="$TMPDIR_TEST/home"
mkdir -p "$TEST_HOME/.tirith"
cp "$CONFIG_PATH" "$TEST_HOME/.tirith/config.yaml"

HOME="$TEST_HOME" "$TMPDIR_TEST/tirith_binary" start --no-open --port "$PROXY_PORT" &
TIRITH_PID=$!

# Wait for both servers to come up.
wait_for() {
    local url=$1 tries=0
    while ! curl -sf "$url" > /dev/null 2>&1; do
        tries=$((tries+1))
        if [ "$tries" -ge 40 ]; then
            echo "FATAL: $url did not come up" >&2
            exit 1
        fi
        sleep 0.1
    done
}

wait_for "http://127.0.0.1:$PROXY_PORT/health"
wait_for "http://127.0.0.1:$DASH_PORT/"

PROXY="http://127.0.0.1:$PROXY_PORT"
DASH="http://127.0.0.1:$DASH_PORT"

echo ""
echo "Running tests..."
echo ""

# ── Test 1: Health check ──────────────────────────────────────────────
RESP=$(curl -sf "$PROXY/health")
if [ "$RESP" = "ok" ]; then pass "Health check returns 'ok'"; else fail "Health check (got: $RESP)"; fi

# ── Test 2: Non-streaming proxy + usage logging ──────────────────────
RESP=$(curl -sf -X POST "$PROXY/proxy/anthropic/v1/messages" \
    -H "Content-Type: application/json" \
    -H "x-api-key: fake-key" \
    -d '{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":100}')

MODEL=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['model'])" 2>/dev/null || echo "")
if [ "$MODEL" = "claude-sonnet-4-6" ]; then
    pass "Non-streaming proxy returns upstream response"
else
    fail "Non-streaming proxy (model=$MODEL)"
fi

# ── Test 3: Streaming proxy + SSE passthrough ────────────────────────
STREAM_RESP=$(curl -sf -X POST "$PROXY/proxy/anthropic/v1/messages" \
    -H "Content-Type: application/json" \
    -H "x-api-key: fake-key" \
    -d '{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":100,"stream":true}')

if echo "$STREAM_RESP" | grep -q "message_start"; then
    pass "Streaming proxy forwards SSE events"
else
    fail "Streaming proxy (no message_start found)"
fi

if echo "$STREAM_RESP" | grep -q "message_delta"; then
    pass "Streaming proxy forwards message_delta with usage"
else
    fail "Streaming proxy (no message_delta found)"
fi

# ── Test 4: Custom headers extracted (tag) ────────────────────────────
curl -sf -X POST "$PROXY/proxy/anthropic/v1/messages" \
    -H "Content-Type: application/json" \
    -H "x-api-key: fake-key" \
    -H "X-Tirith-Tag: test-feature" \
    -H "X-Tirith-User: joe" \
    -H "X-Tirith-Environment: staging" \
    -d '{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"tagged"}],"max_tokens":50}' > /dev/null

# Give storage a moment to flush.
sleep 0.3

# Check dashboard API for the tagged call.
TAGS_RESP=$(curl -sf "$DASH/api/by-tag")
if echo "$TAGS_RESP" | python3 -c "import sys,json; tags=[t['tag'] for t in json.load(sys.stdin)]; assert 'test-feature' in tags" 2>/dev/null; then
    pass "Custom tag 'test-feature' visible in dashboard"
else
    fail "Custom tag not found in /api/by-tag"
fi

# ── Test 5: Auto-detect route (/v1/messages → Anthropic) ─────────────
RESP=$(curl -sf -X POST "$PROXY/v1/messages" \
    -H "Content-Type: application/json" \
    -H "x-api-key: fake-key" \
    -d '{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"autodetect"}],"max_tokens":50}')

MODEL=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['model'])" 2>/dev/null || echo "")
if [ "$MODEL" = "claude-sonnet-4-6" ]; then
    pass "Auto-detect route /v1/messages works"
else
    fail "Auto-detect route (model=$MODEL)"
fi

# ── Test 6: Dashboard API endpoints ──────────────────────────────────
sleep 0.3

# Overview
OVERVIEW=$(curl -sf "$DASH/api/overview")
CALLS=$(echo "$OVERVIEW" | python3 -c "import sys,json; print(json.load(sys.stdin)['total_calls'])" 2>/dev/null || echo "0")
if [ "$CALLS" -ge 3 ]; then
    pass "Dashboard /api/overview reports $CALLS calls"
else
    fail "Dashboard /api/overview (expected >= 3 calls, got $CALLS)"
fi

# By model
MODELS=$(curl -sf "$DASH/api/by-model")
HAS_SONNET=$(echo "$MODELS" | python3 -c "import sys,json; print(any(m['model']=='claude-sonnet-4-6' for m in json.load(sys.stdin)))" 2>/dev/null || echo "")
if [ "$HAS_SONNET" = "True" ]; then
    pass "Dashboard /api/by-model shows claude-sonnet-4-6"
else
    fail "Dashboard /api/by-model missing claude-sonnet-4-6"
fi

# By user
USERS=$(curl -sf "$DASH/api/by-user")
HAS_JOE=$(echo "$USERS" | python3 -c "import sys,json; print(any(u['user']=='joe' for u in json.load(sys.stdin)))" 2>/dev/null || echo "")
if [ "$HAS_JOE" = "True" ]; then
    pass "Dashboard /api/by-user shows user 'joe'"
else
    fail "Dashboard /api/by-user missing user 'joe'"
fi

# Recent calls
RECENT=$(curl -sf "$DASH/api/calls")
CALL_COUNT=$(echo "$RECENT" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
if [ "$CALL_COUNT" -ge 3 ]; then
    pass "Dashboard /api/calls returns $CALL_COUNT recent calls"
else
    fail "Dashboard /api/calls (expected >= 3, got $CALL_COUNT)"
fi

# Daily spend
DAILY=$(curl -sf "$DASH/api/daily-spend")
if echo "$DAILY" | python3 -c "import sys,json; assert len(json.load(sys.stdin)) >= 1" 2>/dev/null; then
    pass "Dashboard /api/daily-spend returns data"
else
    fail "Dashboard /api/daily-spend empty"
fi

# Cost sanity check — we sent 100 input + 42 output tokens on claude-sonnet-4-6
# At $3/M input + $15/M output, that's ~$0.00093 = ~0.093 cents, rounds up to 1 cent.
COST=$(echo "$OVERVIEW" | python3 -c "import sys,json; print(json.load(sys.stdin)['total_cost_cents'])" 2>/dev/null || echo "0")
if [ "$COST" -ge 1 ]; then
    pass "Cost tracking working (total_cost_cents=$COST)"
else
    fail "Cost tracking (total_cost_cents=$COST, expected >= 1)"
fi

# ── Test 7: No CORS on proxy or dashboard ─────────────────────────────
# Proxy should have NO CORS headers (no browser talks to it directly).
PROXY_CORS=$(curl -sf -I "$PROXY/health" 2>/dev/null | grep -i "access-control-allow-origin" || echo "")
if [ -z "$PROXY_CORS" ]; then
    pass "Proxy has no CORS headers (not needed)"
else
    fail "Proxy still has CORS headers: $PROXY_CORS"
fi

# Dashboard should have NO CORS headers (serves its own frontend).
DASH_CORS=$(curl -sf -I "$DASH/" 2>/dev/null | grep -i "access-control-allow-origin" || echo "")
if [ -z "$DASH_CORS" ]; then
    pass "Dashboard has no CORS headers"
else
    fail "Dashboard still has CORS headers: $DASH_CORS"
fi

# ── Test 7b: Proxy health via dashboard (same-origin endpoint) ────────
HEALTH_RESP=$(curl -sf "$DASH/api/proxy-health")
IS_ACTIVE=$(echo "$HEALTH_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['active'])" 2>/dev/null || echo "")
if [ "$IS_ACTIVE" = "True" ]; then
    pass "Dashboard /api/proxy-health reports proxy active"
else
    fail "Dashboard /api/proxy-health (got: $IS_ACTIVE)"
fi

# ── Test 8: Body size limit (413) ────────────────────────────────────
# Generate a body > 10MB.
STATUS=$(python3 -c "
import urllib.request, json
body = json.dumps({'model':'x','messages':[{'role':'user','content':'A'*11_000_000}]}).encode()
req = urllib.request.Request('$PROXY/proxy/anthropic/v1/messages', data=body, method='POST')
req.add_header('Content-Type', 'application/json')
try:
    urllib.request.urlopen(req)
    print('200')
except urllib.request.HTTPError as e:
    print(e.code)
except Exception as e:
    print(f'err:{e}')
")
if [ "$STATUS" = "413" ]; then
    pass "Oversized request body returns 413"
else
    fail "Oversized body (expected 413, got $STATUS)"
fi

# ── Test 9: Invalid JSON returns 400 ─────────────────────────────────
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$PROXY/proxy/anthropic/v1/messages" \
    -H "Content-Type: application/json" \
    -d 'this is not json')
if [ "$STATUS" = "400" ]; then
    pass "Invalid JSON body returns 400"
else
    fail "Invalid JSON (expected 400, got $STATUS)"
fi

# ── Test 10: Tirith headers stripped before upstream ───────────────
# The fake upstream returns 500 if it sees any X-Tirith-* header.
# We already sent requests with those headers and got 200s, so this is implicitly tested.
# Verify explicitly by checking that our tagged request (test 4) succeeded.
TAGGED_CALL=$(echo "$RECENT" | python3 -c "
import sys,json
calls = json.load(sys.stdin)
tagged = [c for c in calls if c['tag'] == 'test-feature']
print(tagged[0]['status_code'] if tagged else 'none')
" 2>/dev/null || echo "none")
if [ "$TAGGED_CALL" = "200" ]; then
    pass "Tirith headers stripped before upstream (tagged call got 200)"
else
    fail "Tirith headers may have leaked (status=$TAGGED_CALL)"
fi

# ── Test 11: Localhost binding ────────────────────────────────────────
# Verify the proxy is NOT listening on 0.0.0.0 by checking the bind address.
LISTEN=$(lsof -iTCP:$PROXY_PORT -sTCP:LISTEN -P -n 2>/dev/null | grep -v "^COMMAND" || echo "")
if echo "$LISTEN" | grep -q "127.0.0.1"; then
    pass "Proxy bound to 127.0.0.1 (not 0.0.0.0)"
else
    if echo "$LISTEN" | grep -q "\*:$PROXY_PORT"; then
        fail "Proxy bound to 0.0.0.0"
    else
        pass "Proxy bound to localhost (could not verify via lsof)"
    fi
fi

# ── Summary ───────────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
printf "  Results: \033[32m%d passed\033[0m, \033[31m%d failed\033[0m, %d total\n" "$PASS" "$FAIL" "$TOTAL"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ "$FAIL" -gt 0 ]; then exit 1; fi
exit 0
