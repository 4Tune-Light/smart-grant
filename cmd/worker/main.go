package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rizky/smart-grant/internal/config"
	"github.com/rizky/smart-grant/internal/notification"
	"github.com/rizky/smart-grant/pkg/database"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.MustLoad("")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, cfg.PostgresConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}
	defer pool.Close()

	rdb, err := database.NewRedisClient(ctx, cfg.RedisConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("Redis is required for worker")
	}
	defer rdb.Close()

	sub := notification.NewSubscriber(rdb, pool)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Msg("Worker started")
		if err := sub.Run(ctx); err != nil {
			log.Error().Err(err).Msg("subscriber stopped")
		}
	}()

	<-sigCh
	log.Info().Msg("shutting down worker")
	cancel()
}
