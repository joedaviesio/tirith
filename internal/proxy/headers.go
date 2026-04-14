package proxy

import (
	"net/http"
	"strings"
)

// safeResponseHeaders is the set of upstream response headers that are safe
// to forward back to the client. Headers not in this set (e.g. Set-Cookie,
// Server, X-Powered-By) are stripped to prevent header injection from a
// compromised upstream.
var safeResponseHeaders = map[string]bool{
	"content-type":              true,
	"content-length":            true,
	"content-encoding":          true,
	"cache-control":             true,
	"date":                      true,
	"vary":                      true,
	"transfer-encoding":         true,
	"x-request-id":              true,
	"request-id":                true,
	"anthropic-ratelimit-limit": true,
	"anthropic-ratelimit-remaining": true,
	"anthropic-ratelimit-reset":     true,
	"retry-after":               true,
	// OpenAI rate-limit / observability headers.
	"openai-model":                     true,
	"openai-organization":              true,
	"openai-processing-ms":             true,
	"openai-version":                   true,
	"x-ratelimit-limit-requests":       true,
	"x-ratelimit-limit-tokens":         true,
	"x-ratelimit-remaining-requests":   true,
	"x-ratelimit-remaining-tokens":     true,
	"x-ratelimit-reset-requests":       true,
	"x-ratelimit-reset-tokens":         true,
}

// copySafeHeaders copies only whitelisted headers from src to dst.
func copySafeHeaders(src, dst http.Header) {
	for key, values := range src {
		if safeResponseHeaders[strings.ToLower(key)] {
			for _, v := range values {
				dst.Add(key, v)
			}
		}
	}
}
