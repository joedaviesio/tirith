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
