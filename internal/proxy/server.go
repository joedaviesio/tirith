package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/joedaviesio/costwatch/internal/config"
	"github.com/joedaviesio/costwatch/internal/pricing"
	"github.com/joedaviesio/costwatch/internal/storage"
)

type Server struct {
	cfg        *config.Config
	store      *storage.Store
	pricer     *pricing.Engine
	httpSrv    *http.Server
	httpClient *http.Client
	logger     *slog.Logger
}

func NewServer(cfg *config.Config, store *storage.Store, pricer *pricing.Engine, logger *slog.Logger) *Server {
	s := &Server{
		cfg:    cfg,
		store:  store,
		pricer: pricer,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

	mux := http.NewServeMux()

	// Explicit provider routing: /proxy/anthropic/...
	mux.HandleFunc("/proxy/anthropic/", s.handleAnthropic)

	// Auto-detect routing: /v1/messages → Anthropic
	mux.HandleFunc("/v1/messages", s.handleAnthropic)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	dashboardOrigin := fmt.Sprintf("http://%s:%d", cfg.Dashboard.Host, cfg.Dashboard.Port)

	s.httpSrv = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port),
		Handler:           corsMiddleware(mux, dashboardOrigin),
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.httpSrv.Addr)
	if err != nil {
		return fmt.Errorf("binding to port %d: %w", s.cfg.Proxy.Port, err)
	}
	s.logger.Info("proxy started", "port", s.cfg.Proxy.Port)
	return s.httpSrv.Serve(ln)
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return s.httpSrv.Addr
}

func corsMiddleware(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
