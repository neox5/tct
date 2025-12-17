// Package server provides HTTP server setup and lifecycle management.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/neox5/tct/internal/logger"
	"github.com/neox5/tct/internal/metrics"
)

// Server manages the HTTP server lifecycle.
type Server struct {
	port   int
	logger *logger.Logger
	mux    *http.ServeMux
}

// New creates a new HTTP server.
func New(port int, log *logger.Logger) *Server {
	return &Server{
		port:   port,
		logger: log,
		mux:    http.NewServeMux(),
	}
}

// RegisterCommonRoutes registers /metrics, /healthz, and /readyz endpoints.
func (s *Server) RegisterCommonRoutes(healthz, readyz http.HandlerFunc) {
	s.mux.Handle("GET /metrics", metrics.Handler())
	s.mux.HandleFunc("GET /healthz", healthz)
	s.mux.HandleFunc("GET /readyz", readyz)
}

// RegisterHandler registers a custom HTTP handler.
func (s *Server) RegisterHandler(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

// Start runs the HTTP server with graceful shutdown support.
// Blocks until the server stops or an error occurs.
func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.mux,
	}

	// Graceful shutdown handler
	go func() {
		<-ctx.Done()
		s.logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("server shutdown error", "error", err)
		}
	}()

	s.logger.Info("starting server", "port", s.port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
