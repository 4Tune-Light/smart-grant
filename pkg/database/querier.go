package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func getTx(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txKey{}).(pgx.Tx)
	return tx
}

type Querier struct {
	pool *pgxpool.Pool
}

func NewQuerier(pool *pgxpool.Pool) *Querier {
	return &Querier{pool: pool}
}

func (q *Querier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx := getTx(ctx); tx != nil {
		return tx.QueryRow(ctx, sql, args...)
	}
	return q.pool.QueryRow(ctx, sql, args...)
}

func (q *Querier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if tx := getTx(ctx); tx != nil {
		return tx.Exec(ctx, sql, args...)
	}
	return q.pool.Exec(ctx, sql, args...)
}

func (q *Querier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx := getTx(ctx); tx != nil {
		return tx.Query(ctx, sql, args...)
	}
	return q.pool.Query(ctx, sql, args...)
}

func (q *Querier) Pool() *pgxpool.Pool {
	return q.pool
}
