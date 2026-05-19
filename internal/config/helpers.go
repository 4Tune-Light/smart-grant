package config

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/rizky/smart-grant/pkg/database"
)

func (c *Config) PostgresConfig() database.PostgresConfig {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.Postgres.User,
		c.Database.Postgres.Password,
		c.Database.Postgres.Host,
		c.Database.Postgres.Port,
		c.Database.Postgres.DBName,
		c.Database.Postgres.SSLMode,
	)

	return database.PostgresConfig{
		DSN:             dsn,
		MaxOpenConns:    c.Database.Postgres.MaxOpenConns,
		MaxIdleConns:    c.Database.Postgres.MaxIdleConns,
		ConnMaxLifetime: c.Database.Postgres.ConnMaxLifetime,
		ConnMaxIdleTime: c.Database.Postgres.ConnMaxIdleTime,
	}
}

func (c *Config) RedisConfig() database.RedisConfig {
	return database.RedisConfig{
		Addr:         fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port),
		Password:     c.Redis.Password,
		DB:           c.Redis.DB,
		PoolSize:     c.Redis.PoolSize,
		MinIdleConns: c.Redis.MinIdleConns,
	}
}

func MustLoad(cfgPath string) *Config {
	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG_PATH")
	}
	cfg, err := Load(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	return cfg
}
