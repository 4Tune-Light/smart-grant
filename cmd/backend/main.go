package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/rizky/smart-grant/internal/config"
	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/server"
	"github.com/rizky/smart-grant/internal/telemetry"
	"github.com/rizky/smart-grant/pkg/database"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := loadConfig()

	ctx := context.Background()

	tp, mp, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName: cfg.OTel.ServiceName + "-backend",
		Environment: cfg.OTel.Environment,
		Endpoint:    cfg.OTel.Endpoint,
		Insecure:    cfg.OTel.Insecure,
		TraceRatio:  cfg.OTel.TraceRatio,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize telemetry")
	}
	defer func() {
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
	}()

	pgCfg := buildPostgresDSN(cfg)
	pgPool, err := database.NewPostgresPool(ctx, pgCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}
	defer pgPool.Close()

	redisCfg := database.RedisConfig{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	}
	if _, err := database.NewRedisClient(ctx, redisCfg); err != nil {
		log.Warn().Err(err).Msg("Redis not available, continuing without it")
	}

	httpSrv := server.NewHTTPServer(
		"backend-http",
		cfg.Server.HTTP.Host,
		cfg.Server.HTTP.Port,
		cfg.Server.HTTP.ReadTimeout,
	)

	registerRoutes(httpSrv.Router())

	grpcSrv, err := server.NewGRPCServer(
		"backend-grpc",
		cfg.Server.GRPC.Host,
		cfg.Server.GRPC.Port,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create gRPC server")
	}

	registerGRPCServices(grpcSrv.Server())

	mgr := server.NewManager(httpSrv, grpcSrv)

	log.Info().Msg("Backend service started")
	if err := mgr.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("server stopped with error")
	}
}

func loadConfig() *config.Config {
	cfgPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	return cfg
}

func buildPostgresDSN(cfg *config.Config) database.PostgresConfig {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.DBName,
		cfg.Database.Postgres.SSLMode,
	)

	return database.PostgresConfig{
		DSN:             dsn,
		MaxOpenConns:    cfg.Database.Postgres.MaxOpenConns,
		MaxIdleConns:    cfg.Database.Postgres.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.Postgres.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.Postgres.ConnMaxIdleTime,
	}
}

func registerRoutes(r chi.Router) {
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS([]string{"*"}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}

func registerGRPCServices(s *grpc.Server) {
}
