package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joedaviesio/costwatch/internal/pricing"
	"github.com/joedaviesio/costwatch/internal/storage"
)

// SSE event types we care about for token tracking.
type sseMessageDelta struct {
	Type  string `json:"type"`
	Usage struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type sseMessageStart struct {
	Type    string `json:"type"`
	Message struct {
		Model string         `json:"model"`
		Usage anthropicUsage `json:"usage"`
	} `json:"message"`
}

func (s *Server) handleAnthropicStreaming(
	w http.ResponseWriter,
	upstreamReq *http.Request,
	start time.Time,
	tag, userTag, sessionTag, environment, endpoint string,
) {
	resp, err := s.httpClient.Do(upstreamReq)
	if err != nil {
		s.logger.Error("upstream streaming request failed", "error", err)
		http.Error(w, "upstream request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers.
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.logger.Error("response writer does not support flushing")
		return
	}

	// Track usage across SSE events.
	var model string
	var inputTokens, outputTokens, cacheReadTokens, cacheWriteTokens int

	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer for large SSE events.
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// Forward line to client immediately.
		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()

		// Parse SSE data lines for usage tracking.
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			continue
		}

		var event struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "message_start":
			var msg sseMessageStart
			if err := json.Unmarshal([]byte(data), &msg); err == nil {
				model = msg.Message.Model
				inputTokens = msg.Message.Usage.InputTokens
				cacheReadTokens = msg.Message.Usage.CacheReadTokens
				cacheWriteTokens = msg.Message.Usage.CacheWriteTokens
			}
		case "message_delta":
			var delta sseMessageDelta
			if err := json.Unmarshal([]byte(data), &delta); err == nil {
				outputTokens = delta.Usage.OutputTokens
			}
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Warn("error reading SSE stream", "error", err)
	}

	// Send final newline after stream.
	fmt.Fprint(w, "\n")
	flusher.Flush()

	latencyMs := int(time.Since(start).Milliseconds())

	// Calculate cost.
	costCents := 0
	if model != "" {
		usage := pricing.TokenUsage{
			InputTokens:      inputTokens,
			OutputTokens:     outputTokens,
			CacheReadTokens:  cacheReadTokens,
			CacheWriteTokens: cacheWriteTokens,
		}
		if c, err := s.pricer.CalculateCostCents("anthropic", model, usage); err == nil {
			costCents = c
		} else {
			s.logger.Warn("cost calculation failed", "model", model, "error", err)
		}
	}

	// Log to storage.
	call := &storage.APICall{
		ID:               uuid.New().String(),
		Timestamp:        start,
		Provider:         "anthropic",
		Model:            model,
		InputTokens:      inputTokens,
		OutputTokens:     outputTokens,
		CacheReadTokens:  cacheReadTokens,
		CacheWriteTokens: cacheWriteTokens,
		CostCents:        costCents,
		LatencyMs:        latencyMs,
		StatusCode:       resp.StatusCode,
		Streaming:        true,
		Tag:              tag,
		UserTag:          userTag,
		SessionTag:       sessionTag,
		Environment:      environment,
		Endpoint:         endpoint,
	}

	if err := s.store.Insert(call); err != nil {
		s.logger.Error("failed to log streaming api call", "error", err)
	} else {
		s.logger.Info("logged streaming api call",
			"model", call.Model,
			"input_tokens", call.InputTokens,
			"output_tokens", call.OutputTokens,
			"cost_cents", call.CostCents,
			"latency_ms", call.LatencyMs,
			"tag", call.Tag,
		)
	}
}

