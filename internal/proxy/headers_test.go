package proxy

import (
	"net/http"
	"testing"
)

func TestCopySafeHeaders_AllowsWhitelisted(t *testing.T) {
	src := http.Header{}
	src.Set("Content-Type", "application/json")
	src.Set("Content-Length", "42")
	src.Set("Content-Encoding", "gzip")
	src.Set("Cache-Control", "no-cache")
	src.Set("X-Request-Id", "abc-123")
	src.Set("Retry-After", "30")

	dst := http.Header{}
	copySafeHeaders(src, dst)

	for _, key := range []string{"Content-Type", "Content-Length", "Content-Encoding", "Cache-Control", "X-Request-Id", "Retry-After"} {
		if dst.Get(key) == "" {
			t.Errorf("expected header %s to be copied, but it was not", key)
		}
	}
}

func TestCopySafeHeaders_BlocksDangerous(t *testing.T) {
	src := http.Header{}
	src.Set("Set-Cookie", "session=abc123; HttpOnly")
	src.Set("Server", "nginx")
	src.Set("X-Powered-By", "Go")
	src.Set("X-Custom-Malicious", "evil")
	src.Set("Authorization", "Bearer leaked-token")

	dst := http.Header{}
	copySafeHeaders(src, dst)

	for _, key := range []string{"Set-Cookie", "Server", "X-Powered-By", "X-Custom-Malicious", "Authorization"} {
		if dst.Get(key) != "" {
			t.Errorf("header %s should have been stripped, but got %q", key, dst.Get(key))
		}
	}
}

func TestCopySafeHeaders_PreservesMultipleValues(t *testing.T) {
	src := http.Header{}
	src.Add("Vary", "Accept-Encoding")
	src.Add("Vary", "Origin")

	dst := http.Header{}
	copySafeHeaders(src, dst)

	values := dst.Values("Vary")
	if len(values) != 2 {
		t.Errorf("Vary values = %d, want 2", len(values))
	}
}

func TestCopySafeHeaders_EmptyInput(t *testing.T) {
	dst := http.Header{}
	copySafeHeaders(http.Header{}, dst)

	if len(dst) != 0 {
		t.Errorf("expected empty dst, got %d headers", len(dst))
	}
}

func TestCopySafeHeaders_AnthropicRateLimitHeaders(t *testing.T) {
	src := http.Header{}
	src.Set("Anthropic-Ratelimit-Limit", "1000")
	src.Set("Anthropic-Ratelimit-Remaining", "999")
	src.Set("Anthropic-Ratelimit-Reset", "2026-04-12T00:00:00Z")

	dst := http.Header{}
	copySafeHeaders(src, dst)

	for _, key := range []string{"Anthropic-Ratelimit-Limit", "Anthropic-Ratelimit-Remaining", "Anthropic-Ratelimit-Reset"} {
		if dst.Get(key) == "" {
			t.Errorf("expected Anthropic rate limit header %s to be copied", key)
		}
	}
}
