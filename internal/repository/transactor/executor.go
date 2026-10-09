package transactor

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueryExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func GetExecutor(ctx context.Context, pool *pgxpool.Pool) QueryExecutor {
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		return tx
	}

	return pool
}

func IsTx(ctx context.Context) bool {
	_, ok := ctx.Value(TxKey{}).(pgx.Tx)

	return ok
}
