package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rizky/smart-grant/internal/config"
	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/server"
	"github.com/rizky/smart-grant/internal/telemetry"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.MustLoad("")

	ctx := context.Background()

	tp, mp, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName: cfg.OTel.ServiceName + "-gateway",
		Environment: cfg.OTel.Environment,
		Endpoint:    cfg.OTel.Endpoint,
		Insecure:    cfg.OTel.Insecure,
		TraceRatio:  cfg.OTel.TraceRatio,
	})
	if err != nil {
		log.Warn().Err(err).Msg("telemetry not available")
	} else {
		defer func() { _ = tp.Shutdown(ctx); _ = mp.Shutdown(ctx) }()
	}

	httpSrv := server.NewHTTPServer("gateway-http",
		cfg.Gateway.HTTP.Host, cfg.Gateway.HTTP.Port, cfg.Gateway.HTTP.ReadTimeout)

	registerRoutes(httpSrv.Router(), cfg)

	mgr := server.NewManager(httpSrv)

	log.Info().Msg("API Gateway started")
	if err := mgr.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("server stopped with error")
	}
}

func registerRoutes(r chi.Router, cfg *config.Config) {
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS([]string{"*"}))
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.OTelHTTP(cfg.OTel.ServiceName + "-gateway"))

	if cfg.Gateway.RateLimit.Enabled {
		rl := middleware.NewRateLimiter(cfg.Gateway.RateLimit.Rate, cfg.Gateway.RateLimit.Burst)
		r.Use(rl.Middleware)
		log.Info().Int("rate", cfg.Gateway.RateLimit.Rate).Int("burst", cfg.Gateway.RateLimit.Burst).Msg("rate limiter enabled")
	}

	if cfg.Gateway.MaxBodySize > 0 {
		r.Use(requestSizeLimit(cfg.Gateway.MaxBodySize))
	}

	r.Get("/health", healthHandler)

	backendURL, err := url.Parse(cfg.Gateway.BackendURL)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid backend URL")
	}
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func requestSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
