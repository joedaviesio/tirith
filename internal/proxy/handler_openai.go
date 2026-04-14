package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joedaviesio/tirith/internal/pricing"
	"github.com/joedaviesio/tirith/internal/storage"
)

// openaiUsage covers both Chat Completions and Responses API shapes.
type openaiUsage struct {
	// Chat Completions
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	PromptTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	// Responses API
	InputTokens        int `json:"input_tokens"`
	OutputTokens       int `json:"output_tokens"`
	InputTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"input_tokens_details"`
}

type openaiResponse struct {
	Model string      `json:"model"`
	Usage openaiUsage `json:"usage"`
}

// normalize reconciles the two API shapes into a single TokenUsage.
func (u openaiUsage) normalize() pricing.TokenUsage {
	in := u.PromptTokens
	out := u.CompletionTokens
	cached := u.PromptTokensDetails.CachedTokens
	if in == 0 && u.InputTokens > 0 {
		in = u.InputTokens
	}
	if out == 0 && u.OutputTokens > 0 {
		out = u.OutputTokens
	}
	if cached == 0 && u.InputTokensDetails.CachedTokens > 0 {
		cached = u.InputTokensDetails.CachedTokens
	}
	// OpenAI reports cached tokens as a subset of prompt/input tokens; subtract
	// so we don't double-count when pricing cache reads separately.
	if cached > 0 && in >= cached {
		in -= cached
	}
	return pricing.TokenUsage{
		InputTokens:     in,
		OutputTokens:    out,
		CacheReadTokens: cached,
	}
}

func (s *Server) handleOpenAI(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	tag := r.Header.Get("X-Tirith-Tag")
	userTag := r.Header.Get("X-Tirith-User")
	sessionTag := r.Header.Get("X-Tirith-Session")
	environment := r.Header.Get("X-Tirith-Environment")
	if environment == "" {
		environment = "default"
	}

	for key := range r.Header {
		if strings.HasPrefix(strings.ToLower(key), "x-tirith-") {
			r.Header.Del(key)
		}
	}

	const maxBodySize = 10 << 20
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
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
	}

	isStreaming := false
	if stream, ok := reqBody["stream"].(bool); ok && stream {
		isStreaming = true
	}

	// Ensure SSE streams report usage (OpenAI requires opt-in).
	if isStreaming {
		opts, _ := reqBody["stream_options"].(map[string]interface{})
		if opts == nil {
			opts = map[string]interface{}{}
		}
		if _, present := opts["include_usage"]; !present {
			opts["include_usage"] = true
			reqBody["stream_options"] = opts
			if re, err := json.Marshal(reqBody); err == nil {
				bodyBytes = re
			}
		}
	}

	// Strip our proxy prefix and, if the client didn't include /v1 or
	// /responses, prepend /v1 so upstream resolves.
	upstreamPath := strings.TrimPrefix(r.URL.Path, "/proxy/openai")
	if !strings.HasPrefix(upstreamPath, "/v1") && !strings.HasPrefix(upstreamPath, "/responses") {
		upstreamPath = "/v1" + upstreamPath
	}
	upstreamURL := s.cfg.Providers.OpenAI.Upstream + upstreamPath
	if r.URL.RawQuery != "" {
		upstreamURL += "?" + r.URL.RawQuery
	}

	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(bodyBytes))
	if err != nil {
		s.logger.Error("creating upstream request", "error", err)
		http.Error(w, "internal proxy error", http.StatusBadGateway)
		return
	}

	for key, values := range r.Header {
		// Content-Length will be set automatically and may have changed
		// if we rewrote the body above.
		if strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, v := range values {
			upstreamReq.Header.Add(key, v)
		}
	}

	if isStreaming {
		s.handleOpenAIStreaming(w, upstreamReq, start, tag, userTag, sessionTag, environment, upstreamPath)
		return
	}

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

	parseBody := respBody
	if resp.Header.Get("Content-Encoding") == "gzip" {
		if gr, err := gzip.NewReader(bytes.NewReader(respBody)); err == nil {
			if decompressed, err := io.ReadAll(gr); err == nil {
				parseBody = decompressed
			}
			gr.Close()
		}
	}

	var apiResp openaiResponse
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(parseBody, &apiResp); err != nil {
			s.logger.Warn("failed to parse openai response for usage", "error", err)
		}
	}

	usage := apiResp.Usage.normalize()
	costCents := 0
	if apiResp.Model != "" {
		if c, err := s.pricer.CalculateCostCents("openai", apiResp.Model, usage); err == nil {
			costCents = c
		} else {
			s.logger.Warn("cost calculation failed", "model", apiResp.Model, "error", err)
		}
	}

	call := &storage.APICall{
		ID:              uuid.New().String(),
		Timestamp:       start,
		Provider:        "openai",
		Model:           apiResp.Model,
		InputTokens:     usage.InputTokens,
		OutputTokens:    usage.OutputTokens,
		CacheReadTokens: usage.CacheReadTokens,
		CostCents:       costCents,
		LatencyMs:       latencyMs,
		StatusCode:      resp.StatusCode,
		Streaming:       false,
		Tag:             tag,
		UserTag:         userTag,
		SessionTag:      sessionTag,
		Environment:     environment,
		Endpoint:        upstreamPath,
	}

	if err := s.store.Insert(call); err != nil {
		s.logger.Error("failed to log api call", "error", err)
	} else {
		s.logger.Info("logged api call",
			"provider", "openai",
			"model", call.Model,
			"input_tokens", call.InputTokens,
			"output_tokens", call.OutputTokens,
			"cost_cents", call.CostCents,
			"latency_ms", call.LatencyMs,
			"tag", call.Tag,
		)
	}

	copySafeHeaders(resp.Header, w.Header())
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}
