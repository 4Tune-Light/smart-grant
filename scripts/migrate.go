package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/migrate.go [up|down]")
		os.Exit(1)
	}

	action := os.Args[1]

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/smart_grant?sslmode=disable"
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	switch action {
	case "up":
		runMigrations(ctx, pool, "up")
	case "down":
		runMigrations(ctx, pool, "down")
	default:
		fmt.Printf("unknown action: %s (use 'up' or 'down')\n", action)
		os.Exit(1)
	}
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, direction string) {
	migrationsDir := "migrations"
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read migrations directory")
	}

	files := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files[entry.Name()] = filepath.Join(migrationsDir, entry.Name())
	}

	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if direction == "down" {
		sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	}

	for _, name := range keys {
		if !strings.HasSuffix(name, "."+direction+".sql") {
			continue
		}

		content, err := os.ReadFile(files[name])
		if err != nil {
			log.Fatal().Err(err).Str("file", name).Msg("failed to read migration file")
		}

		sql := strings.TrimSpace(string(content))
		if sql == "" {
			continue
		}

		_, err = pool.Exec(ctx, sql)
		if err != nil {
			log.Fatal().Err(err).Str("file", name).Msg("migration failed")
		}

		log.Info().Str("file", name).Msg("migration applied")
	}

	log.Info().Str("direction", direction).Msg("all migrations completed")
}
