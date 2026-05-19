package main

import (
	"context"
	"fmt"
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

	cfgPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.DBName,
		cfg.Database.Postgres.SSLMode,
	)

	pool, err := database.NewPostgresPool(ctx, database.PostgresConfig{
		DSN:             dsn,
		MaxOpenConns:    cfg.Database.Postgres.MaxOpenConns,
		MaxIdleConns:    cfg.Database.Postgres.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.Postgres.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.Postgres.ConnMaxIdleTime,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}
	defer pool.Close()

	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	rdb, err := database.NewRedisClient(ctx, database.RedisConfig{
		Addr:         redisAddr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Redis is required for worker")
	}
	defer rdb.Close()

	sub := notification.NewSubscriber(rdb, pool)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Msg("Worker started — listening for notification events")
		if err := sub.Run(ctx); err != nil {
			log.Error().Err(err).Msg("subscriber stopped with error")
		}
	}()

	<-sigCh
	log.Info().Msg("shutting down worker")
	cancel()
}
