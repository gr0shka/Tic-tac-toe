package postgres

import (
	"context"
	"fmt"

	"github.com/gr0shka/Tic-tac-toe/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func NewClient(lc fx.Lifecycle, cf config.Postgres) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cf.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = 30
	poolConfig.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}
