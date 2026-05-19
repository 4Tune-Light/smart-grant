package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestPool(t *testing.T, ctx context.Context, dsn string) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	return pool
}

func runMigration(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email         VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		name          VARCHAR(255) NOT NULL,
		role          VARCHAR(50)  NOT NULL CHECK (role IN ('admin', 'reviewer', 'applicant')),
		is_active     BOOLEAN      NOT NULL DEFAULT true,
		created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
		updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
	);`
	_, err := pool.Exec(context.Background(), schema)
	if err != nil {
		t.Fatal(fmt.Errorf("migration: %w", err))
	}
}
