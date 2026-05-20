package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) *Transactor {
	return &Transactor{pool: pool}
}

func (t *Transactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) (err error) {
	if t == nil || t.pool == nil {
		return fn(ctx)
	}

	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		}
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	ctx = WithTx(ctx, tx)

	if err = fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
