package proxy

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joedaviesio/tirith/internal/storage"
)

// openaiChatChunk is a chat.completions streaming chunk.
type openaiChatChunk struct {
	Model string       `json:"model"`
	Usage *openaiUsage `json:"usage"`
}

// openaiResponsesStreamEvent covers Responses API SSE events.
// The final "response.completed" event contains the full response including usage.
type openaiResponsesStreamEvent struct {
	Type     string `json:"type"`
	Response *struct {
		Model string      `json:"model"`
		Usage openaiUsage `json:"usage"`
	} `json:"response"`
}

func (s *Server) handleOpenAIStreaming(
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

	copySafeHeaders(resp.Header, w.Header())
	w.WriteHeader(resp.StatusCode)

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.logger.Error("response writer does not support flushing")
		return
	}

	var body io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			s.logger.Error("failed to create gzip reader for streaming response", "error", err)
			return
		}
		defer gr.Close()
		body = gr
	}

	var model string
	var usage openaiUsage

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			continue
		}

		// Try chat.completions chunk (flat usage on the chunk).
		var chat openaiChatChunk
		if err := json.Unmarshal([]byte(data), &chat); err == nil {
			if chat.Model != "" {
				model = chat.Model
			}
			if chat.Usage != nil {
				usage = *chat.Usage
				continue
			}
		}

		// Try Responses API event envelope.
		var ev openaiResponsesStreamEvent
		if err := json.Unmarshal([]byte(data), &ev); err == nil && ev.Response != nil {
			if ev.Response.Model != "" {
				model = ev.Response.Model
			}
			// Only overwrite usage when the event actually reports it
			// (response.completed), so partial deltas don't clobber.
			if ev.Type == "response.completed" ||
				ev.Response.Usage.InputTokens > 0 ||
				ev.Response.Usage.OutputTokens > 0 {
				usage = ev.Response.Usage
			}
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Warn("error reading SSE stream", "error", err)
	}

	fmt.Fprint(w, "\n")
	flusher.Flush()

	latencyMs := int(time.Since(start).Milliseconds())

	normalized := usage.normalize()
	costCents := 0
	if model != "" {
		if c, err := s.pricer.CalculateCostCents("openai", model, normalized); err == nil {
			costCents = c
		} else {
			s.logger.Warn("cost calculation failed", "model", model, "error", err)
		}
	}

	call := &storage.APICall{
		ID:              uuid.New().String(),
		Timestamp:       start,
		Provider:        "openai",
		Model:           model,
		InputTokens:     normalized.InputTokens,
		OutputTokens:    normalized.OutputTokens,
		CacheReadTokens: normalized.CacheReadTokens,
		CostCents:       costCents,
		LatencyMs:       latencyMs,
		StatusCode:      resp.StatusCode,
		Streaming:       true,
		Tag:             tag,
		UserTag:         userTag,
		SessionTag:      sessionTag,
		Environment:     environment,
		Endpoint:        endpoint,
	}

	if err := s.store.Insert(call); err != nil {
		s.logger.Error("failed to log streaming api call", "error", err)
	} else {
		s.logger.Info("logged streaming api call",
			"provider", "openai",
			"model", call.Model,
			"input_tokens", call.InputTokens,
			"output_tokens", call.OutputTokens,
			"cost_cents", call.CostCents,
			"latency_ms", call.LatencyMs,
			"tag", call.Tag,
		)
	}
}
