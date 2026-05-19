package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rizky/smart-grant/internal/config"
	"github.com/rizky/smart-grant/internal/server"
	"github.com/rizky/smart-grant/internal/telemetry"
	"github.com/rizky/smart-grant/pkg/database"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.MustLoad("")

	ctx := context.Background()

	tp, mp, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName: cfg.OTel.ServiceName + "-backend",
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

	pool, err := database.NewPostgresPool(ctx, cfg.PostgresConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}
	defer pool.Close()

	rdb, err := database.NewRedisClient(ctx, cfg.RedisConfig())
	if err != nil {
		log.Warn().Err(err).Msg("Redis not available")
		rdb = nil
	}

	httpSrv := server.NewHTTPServer("backend-http",
		cfg.Server.HTTP.Host, cfg.Server.HTTP.Port, cfg.Server.HTTP.ReadTimeout)

	grpcSrv, err := server.NewGRPCServer("backend-grpc", cfg.Server.GRPC.Host, cfg.Server.GRPC.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create gRPC server")
	}

	server.RegisterRoutes(httpSrv.Router(), cfg, pool, rdb)
	server.RegisterGRPC(grpcSrv.Server(), pool)

	mgr := server.NewManager(httpSrv, grpcSrv)

	log.Info().Msg("Backend service started")
	if err := mgr.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("server stopped with error")
	}
}
