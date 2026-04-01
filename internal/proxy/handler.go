package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joedaviesio/tirith/internal/pricing"
	"github.com/joedaviesio/tirith/internal/storage"
)

// anthropicUsage represents the usage object in an Anthropic API response.
type anthropicUsage struct {
	InputTokens      int `json:"input_tokens"`
	OutputTokens     int `json:"output_tokens"`
	CacheReadTokens  int `json:"cache_read_input_tokens"`
	CacheWriteTokens int `json:"cache_creation_input_tokens"`
}

type anthropicResponse struct {
	Model string         `json:"model"`
	Usage anthropicUsage `json:"usage"`
}

func (s *Server) handleAnthropic(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Extract Tirith headers before stripping them.
	tag := r.Header.Get("X-Tirith-Tag")
	userTag := r.Header.Get("X-Tirith-User")
	sessionTag := r.Header.Get("X-Tirith-Session")
	environment := r.Header.Get("X-Tirith-Environment")
	if environment == "" {
		environment = "default"
	}

	// Strip Tirith headers.
	for key := range r.Header {
		if strings.HasPrefix(strings.ToLower(key), "x-tirith-") {
			r.Header.Del(key)
		}
	}

	// Read request body with a 10MB size limit to prevent memory exhaustion.
	const maxBodySize = 10 << 20 // 10MB
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize+1))
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()
	if len(bodyBytes) > maxBodySize {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var reqBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	isStreaming := false
	if stream, ok := reqBody["stream"].(bool); ok && stream {
		isStreaming = true
	}

	// Build upstream URL.
	upstreamPath := r.URL.Path
	upstreamPath = strings.TrimPrefix(upstreamPath, "/proxy/anthropic")
	upstreamURL := s.cfg.Providers.Anthropic.Upstream + upstreamPath

	// Create upstream request.
	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(bodyBytes))
	if err != nil {
		s.logger.Error("creating upstream request", "error", err)
		http.Error(w, "internal proxy error", http.StatusBadGateway)
		return
	}

	// Copy headers from original request.
	for key, values := range r.Header {
		for _, v := range values {
			upstreamReq.Header.Add(key, v)
		}
	}

	if isStreaming {
		s.handleAnthropicStreaming(w, upstreamReq, start, tag, userTag, sessionTag, environment, upstreamPath)
		return
	}

	// Non-streaming: forward request and read full response.
	resp, err := s.httpClient.Do(upstreamReq)
	if err != nil {
		s.logger.Error("upstream request failed", "error", err)
		http.Error(w, "upstream request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("reading upstream response", "error", err)
		http.Error(w, "failed to read upstream response", http.StatusBadGateway)
		return
	}

	latencyMs := int(time.Since(start).Milliseconds())

	// Parse usage from response.
	var apiResp anthropicResponse
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			s.logger.Warn("failed to parse response for usage", "error", err)
		}
	}

	// Calculate cost.
	costCents := 0
	if apiResp.Model != "" {
		usage := pricing.TokenUsage{
			InputTokens:      apiResp.Usage.InputTokens,
			OutputTokens:     apiResp.Usage.OutputTokens,
			CacheReadTokens:  apiResp.Usage.CacheReadTokens,
			CacheWriteTokens: apiResp.Usage.CacheWriteTokens,
		}
		if c, err := s.pricer.CalculateCostCents("anthropic", apiResp.Model, usage); err == nil {
			costCents = c
		} else {
			s.logger.Warn("cost calculation failed", "model", apiResp.Model, "error", err)
		}
	}

	// Log to storage.
	call := &storage.APICall{
		ID:               uuid.New().String(),
		Timestamp:        start,
		Provider:         "anthropic",
		Model:            apiResp.Model,
		InputTokens:      apiResp.Usage.InputTokens,
		OutputTokens:     apiResp.Usage.OutputTokens,
		CacheReadTokens:  apiResp.Usage.CacheReadTokens,
		CacheWriteTokens: apiResp.Usage.CacheWriteTokens,
		CostCents:        costCents,
		LatencyMs:        latencyMs,
		StatusCode:       resp.StatusCode,
		Streaming:        false,
		Tag:              tag,
		UserTag:          userTag,
		SessionTag:       sessionTag,
		Environment:      environment,
		Endpoint:         upstreamPath,
	}

	if err := s.store.Insert(call); err != nil {
		s.logger.Error("failed to log api call", "error", err)
	} else {
		s.logger.Info("logged api call",
			"model", call.Model,
			"input_tokens", call.InputTokens,
			"output_tokens", call.OutputTokens,
			"cost_cents", call.CostCents,
			"latency_ms", call.LatencyMs,
			"tag", call.Tag,
		)
	}

	// Return response to client unmodified.
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}
