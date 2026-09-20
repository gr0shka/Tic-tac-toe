package postgres

import (
	"context"

	"github.com/gr0shka/Tic-tac-toe/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewClient(ctx context.Context, cf config.Postgres) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cf.ConnectionString())
	if err != nil {
		return nil, err
	}

	return pool, nil
}
