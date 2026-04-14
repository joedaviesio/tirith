package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joedaviesio/tirith/internal/config"
	"github.com/joedaviesio/tirith/internal/pricing"
	"github.com/joedaviesio/tirith/internal/storage"
)

// setupProxy creates a proxy server backed by a fake upstream and temp SQLite DB.
// Returns the proxy handler and a cleanup function.
func setupProxy(t *testing.T, upstreamHandler http.HandlerFunc) (*Server, *storage.Store) {
	t.Helper()

	dir := t.TempDir()
	store, err := storage.New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	upstream := httptest.NewServer(upstreamHandler)
	t.Cleanup(upstream.Close)

	pricer, err := pricing.New()
	if err != nil {
		t.Fatalf("creating pricer: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Providers.Anthropic.Upstream = upstream.URL

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewServer(cfg, store, pricer, logger), store
}

func fakeAnthropicUpstream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check that Tirith headers are stripped.
		for key := range r.Header {
			if strings.HasPrefix(strings.ToLower(key), "x-tirith-") {
				w.WriteHeader(500)
				fmt.Fprintf(w, `{"error":"tirith header leaked: %s"}`, key)
				return
			}
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)

		if stream, ok := req["stream"].(bool); ok && stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(200)

			start := `{"type":"message_start","message":{"model":"claude-sonnet-4-6","usage":{"input_tokens":50,"output_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`
			fmt.Fprintf(w, "event: message_start\ndata: %s\n\n", start)

			delta := `{"type":"message_delta","usage":{"output_tokens":25}}`
			fmt.Fprintf(w, "event: message_delta\ndata: %s\n\n", delta)

			fmt.Fprint(w, "data: [DONE]\n\n")
			return
		}

		resp := `{"id":"msg_test","type":"message","model":"claude-sonnet-4-6","role":"assistant","content":[{"type":"text","text":"Hello"}],"usage":{"input_tokens":100,"output_tokens":42,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "evil=1")
		w.Header().Set("Server", "evil-server")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(resp))
	}
}

func TestProxyNonStreaming(t *testing.T) {
	srv, store := setupProxy(t, fakeAnthropicUpstream())

	body := `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":100}`
	req := httptest.NewRequest("POST", "/proxy/anthropic/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", "test-key")

	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["model"] != "claude-sonnet-4-6" {
		t.Errorf("model = %v, want claude-sonnet-4-6", resp["model"])
	}

	// Verify call was logged to storage.
	calls, err := store.GetRecentCalls(10)
	if err != nil {
		t.Fatalf("GetRecentCalls: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("logged calls = %d, want 1", len(calls))
	}
	if calls[0].Model != "claude-sonnet-4-6" {
		t.Errorf("logged model = %s, want claude-sonnet-4-6", calls[0].Model)
	}
	if calls[0].InputTokens != 100 {
		t.Errorf("logged input_tokens = %d, want 100", calls[0].InputTokens)
	}
	if calls[0].OutputTokens != 42 {
		t.Errorf("logged output_tokens = %d, want 42", calls[0].OutputTokens)
	}
	if calls[0].CostCents < 1 {
		t.Errorf("logged cost_cents = %d, want >= 1", calls[0].CostCents)
	}
}

func TestProxyStreaming(t *testing.T) {
	srv, store := setupProxy(t, fakeAnthropicUpstream())

	body := `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":100,"stream":true}`
	req := httptest.NewRequest("POST", "/proxy/anthropic/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	respBody := w.Body.String()
	if !strings.Contains(respBody, "message_start") {
		t.Error("response missing message_start event")
	}
	if !strings.Contains(respBody, "message_delta") {
		t.Error("response missing message_delta event")
	}

	// Verify call was logged.
	calls, err := store.GetRecentCalls(10)
	if err != nil {
		t.Fatalf("GetRecentCalls: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("logged calls = %d, want 1", len(calls))
	}
	if !calls[0].Streaming {
		t.Error("logged call should be streaming=true")
	}
	if calls[0].InputTokens != 50 {
		t.Errorf("logged input_tokens = %d, want 50", calls[0].InputTokens)
	}
	if calls[0].OutputTokens != 25 {
		t.Errorf("logged output_tokens = %d, want 25", calls[0].OutputTokens)
	}
}

func TestTirithHeadersStripped(t *testing.T) {
	srv, store := setupProxy(t, fakeAnthropicUpstream())

	body := `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":100}`
	req := httptest.NewRequest("POST", "/proxy/anthropic/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tirith-Tag", "my-feature")
	req.Header.Set("X-Tirith-User", "alice")
	req.Header.Set("X-Tirith-Session", "sess-123")
	req.Header.Set("X-Tirith-Environment", "staging")

	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	// If upstream saw Tirith headers, it returns 500.
	if w.Code != 200 {
		t.Fatalf("upstream saw Tirith headers (status=%d): %s", w.Code, w.Body.String())
	}

	// Verify headers were captured in the log.
	calls, _ := store.GetRecentCalls(10)
	if len(calls) != 1 {
		t.Fatalf("logged calls = %d, want 1", len(calls))
	}
	if calls[0].Tag != "my-feature" {
		t.Errorf("tag = %q, want my-feature", calls[0].Tag)
	}
	if calls[0].UserTag != "alice" {
		t.Errorf("user_tag = %q, want alice", calls[0].UserTag)
	}
	if calls[0].Environment != "staging" {
		t.Errorf("environment = %q, want staging", calls[0].Environment)
	}
}

func TestResponseHeadersFiltered(t *testing.T) {
	srv, _ := setupProxy(t, fakeAnthropicUpstream())

	body := `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":100}`
	req := httptest.NewRequest("POST", "/proxy/anthropic/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// Set-Cookie and Server headers from upstream should be stripped.
	if cookie := w.Header().Get("Set-Cookie"); cookie != "" {
		t.Errorf("Set-Cookie header leaked: %q", cookie)
	}
	if server := w.Header().Get("Server"); server != "" {
		t.Errorf("Server header leaked: %q", server)
	}

	// Content-Type should pass through.
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestAutoDetectRoute(t *testing.T) {
	srv, _ := setupProxy(t, fakeAnthropicUpstream())

	body := `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":100}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("auto-detect route returned %d, want 200", w.Code)
	}
}

func TestInvalidJSON(t *testing.T) {
	srv, _ := setupProxy(t, fakeAnthropicUpstream())

	req := httptest.NewRequest("POST", "/proxy/anthropic/v1/messages", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400 for invalid JSON", w.Code)
	}
}

func TestOversizedBody(t *testing.T) {
	srv, _ := setupProxy(t, fakeAnthropicUpstream())

	// Create a body > 10MB.
	bigBody := `{"model":"x","messages":[{"role":"user","content":"` + strings.Repeat("A", 11*1024*1024) + `"}]}`
	req := httptest.NewRequest("POST", "/proxy/anthropic/v1/messages", strings.NewReader(bigBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	if w.Code != 413 {
		t.Errorf("status = %d, want 413 for oversized body", w.Code)
	}
}

func TestHealthCheck(t *testing.T) {
	srv, _ := setupProxy(t, fakeAnthropicUpstream())

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("health check status = %d, want 200", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("health check body = %q, want ok", w.Body.String())
	}
}

func TestQueryStringPreserved(t *testing.T) {
	var receivedQuery string
	upstream := func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"model":"claude-sonnet-4-6","usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`))
	}

	srv, _ := setupProxy(t, upstream)

	body := `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":100}`
	req := httptest.NewRequest("POST", "/proxy/anthropic/v1/messages?beta=true&version=2", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if receivedQuery != "beta=true&version=2" {
		t.Errorf("upstream received query = %q, want %q", receivedQuery, "beta=true&version=2")
	}
}

func TestDefaultEnvironment(t *testing.T) {
	srv, store := setupProxy(t, fakeAnthropicUpstream())

	body := `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}],"max_tokens":100}`
	req := httptest.NewRequest("POST", "/proxy/anthropic/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No X-Tirith-Environment header — should default to "default".

	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, req)

	calls, _ := store.GetRecentCalls(10)
	if len(calls) != 1 {
		t.Fatalf("logged calls = %d, want 1", len(calls))
	}
	if calls[0].Environment != "default" {
		t.Errorf("environment = %q, want default", calls[0].Environment)
	}
}

// Suppress unused import warning for os package (used in other test files in this package).
var _ = os.DevNull
