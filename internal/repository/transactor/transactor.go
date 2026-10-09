package transactor

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TxKey struct{}

type transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) *transactor {
	return &transactor{pool: pool}
}

func (t transactor) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txCtx := context.WithValue(ctx, TxKey{}, tx)

	if err = fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
