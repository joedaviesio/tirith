package dashboard

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/joedaviesio/tirith/internal/storage"
)

//go:embed all:frontend
var frontendFS embed.FS

type Server struct {
	store   *storage.Store
	httpSrv *http.Server
	logger  *slog.Logger
	port    int
}

func NewServer(host string, port int, store *storage.Store, logger *slog.Logger) *Server {
	s := &Server{
		store:  store,
		logger: logger,
		port:   port,
	}

	mux := http.NewServeMux()

	// API routes.
	mux.HandleFunc("/api/overview", s.handleOverview)
	mux.HandleFunc("/api/by-model", s.handleByModel)
	mux.HandleFunc("/api/by-tag", s.handleByTag)
	mux.HandleFunc("/api/daily-spend", s.handleDailySpend)
	mux.HandleFunc("/api/daily-by-model", s.handleDailyByModel)
	mux.HandleFunc("/api/by-user", s.handleByUser)
	mux.HandleFunc("/api/calls", s.handleCalls)
	mux.HandleFunc("/api/models", s.handleModels)

	// Serve embedded frontend.
	frontendContent, _ := fs.Sub(frontendFS, "frontend")
	mux.Handle("/", http.FileServer(http.FS(frontendContent)))

	s.httpSrv = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.httpSrv.Addr)
	if err != nil {
		return fmt.Errorf("binding dashboard to port %d: %w", s.port, err)
	}
	s.logger.Info("dashboard started", "port", s.port)
	return s.httpSrv.Serve(ln)
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpSrv.Shutdown(ctx)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) getFilter(r *http.Request) storage.ReportFilter {
	f := storage.ReportFilter{}
	if last := r.URL.Query().Get("last"); last != "" {
		switch last {
		case "1h":
			f.Since = time.Now().Add(-1 * time.Hour)
		case "24h":
			f.Since = time.Now().Add(-24 * time.Hour)
		case "7d":
			f.Since = time.Now().Add(-7 * 24 * time.Hour)
		case "30d":
			f.Since = time.Now().Add(-30 * 24 * time.Hour)
		}
	}
	f.Tag = r.URL.Query().Get("tag")
	f.Model = r.URL.Query().Get("model")
	return f
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	f := s.getFilter(r)
	summary, err := s.store.GetSummary(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{
		"total_cost_cents": summary.TotalCostCents,
		"total_calls":      summary.TotalCalls,
		"total_input":      summary.TotalInput,
		"total_output":     summary.TotalOutput,
		"avg_latency_ms":   summary.AvgLatencyMs,
	})
}

func (s *Server) handleByModel(w http.ResponseWriter, r *http.Request) {
	f := s.getFilter(r)
	models, err := s.store.GetByModel(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(models))
	for i, m := range models {
		result[i] = map[string]interface{}{
			"model":            m.Model,
			"provider":         m.Provider,
			"total_cost_cents": m.TotalCostCents,
			"total_calls":      m.TotalCalls,
			"total_input":      m.TotalInput,
			"total_output":     m.TotalOutput,
			"avg_latency_ms":   m.AvgLatencyMs,
		}
	}
	writeJSON(w, result)
}

func (s *Server) handleByTag(w http.ResponseWriter, r *http.Request) {
	f := s.getFilter(r)
	tags, err := s.store.GetByTag(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(tags))
	for i, t := range tags {
		result[i] = map[string]interface{}{
			"tag":              t.Tag,
			"total_cost_cents": t.TotalCostCents,
			"total_calls":      t.TotalCalls,
			"avg_cost_cents":   t.AvgCostCents,
			"avg_latency_ms":   t.AvgLatencyMs,
		}
	}
	writeJSON(w, result)
}

func (s *Server) handleDailySpend(w http.ResponseWriter, r *http.Request) {
	f := s.getFilter(r)
	daily, err := s.store.GetDailySpend(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(daily))
	for i, d := range daily {
		result[i] = map[string]interface{}{
			"date":             d.Date,
			"total_cost_cents": d.TotalCostCents,
			"total_calls":      d.TotalCalls,
			"total_input":      d.TotalInput,
			"total_output":     d.TotalOutput,
		}
	}
	writeJSON(w, result)
}

func (s *Server) handleDailyByModel(w http.ResponseWriter, r *http.Request) {
	f := s.getFilter(r)
	daily, err := s.store.GetDailyByModel(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(daily))
	for i, d := range daily {
		result[i] = map[string]interface{}{
			"date":             d.Date,
			"model":            d.Model,
			"total_cost_cents": d.TotalCostCents,
			"total_calls":      d.TotalCalls,
			"total_input":      d.TotalInput,
			"total_output":     d.TotalOutput,
		}
	}
	writeJSON(w, result)
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	f := s.getFilter(r)
	models, err := s.store.GetByModel(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]string, len(models))
	for i, m := range models {
		result[i] = m.Model
	}
	writeJSON(w, result)
}

func (s *Server) handleByUser(w http.ResponseWriter, r *http.Request) {
	f := s.getFilter(r)
	users, err := s.store.GetByUser(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(users))
	for i, u := range users {
		result[i] = map[string]interface{}{
			"user":             u.UserTag,
			"total_cost_cents": u.TotalCostCents,
			"total_calls":      u.TotalCalls,
			"first_seen":       u.FirstSeen,
			"last_seen":        u.LastSeen,
		}
	}
	writeJSON(w, result)
}

func (s *Server) handleCalls(w http.ResponseWriter, r *http.Request) {
	calls, err := s.store.GetRecentCalls(50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(calls))
	for i, c := range calls {
		result[i] = map[string]interface{}{
			"id":           c.ID,
			"timestamp":    c.Timestamp.Format(time.RFC3339),
			"provider":     c.Provider,
			"model":        c.Model,
			"input_tokens": c.InputTokens,
			"output_tokens": c.OutputTokens,
			"cost_cents":   c.CostCents,
			"latency_ms":   c.LatencyMs,
			"status_code":  c.StatusCode,
			"streaming":    c.Streaming,
			"tag":          c.Tag,
			"user_tag":     c.UserTag,
			"environment":  c.Environment,
		}
	}
	writeJSON(w, result)
}
