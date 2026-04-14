package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/example/grpc-health-proxy/internal/config"
	"github.com/example/grpc-health-proxy/internal/health"
)

// Server wraps the HTTP server and manages its lifecycle.
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
}

// New creates a new Server with routes wired up from the given config.
func New(cfg *config.Config) (*Server, error) {
	checker, err := health.NewChecker(cfg.GRPCAddress, cfg.GRPCService)
	if err != nil {
		return nil, fmt.Errorf("creating health checker: %w", err)
	}

	handler := health.NewHandler(checker)

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.HTTPPath, handler.ServeHTTP)
	mux.HandleFunc("/livez", livenessHandler)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		cfg:        cfg,
	}, nil
}

// Start begins listening and serving HTTP requests.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func livenessHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
