package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type HTTPServer struct {
	name   string
	server *http.Server
	router chi.Router
}

func NewHTTPServer(name string, host string, port int, timeout time.Duration) *HTTPServer {
	addr := fmt.Sprintf("%s:%d", host, port)

	srv := &HTTPServer{
		name:   name,
		router: chi.NewRouter(),
	}

	srv.server = &http.Server{
		Addr:         addr,
		Handler:      srv.router,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  60 * time.Second,
	}

	return srv
}

func (s *HTTPServer) Router() chi.Router {
	return s.router
}

func (s *HTTPServer) Start() error {
	log.Info().Str("name", s.name).Str("addr", s.server.Addr).Msg("starting HTTP server")
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", s.name, err)
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	log.Info().Str("name", s.name).Msg("shutting down HTTP server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}

func (s *HTTPServer) Name() string {
	return s.name
}
