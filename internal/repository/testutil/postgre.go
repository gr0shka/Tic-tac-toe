package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDB(ctx context.Context) (*pgxpool.Pool, func(), error) {

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),

		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to start postgres: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to connection string: %v", err)
	}

	testPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create test pool: %v", err)
	}

	db := stdlib.OpenDBFromPool(testPool)
	defer db.Close()

	if err = goose.SetDialect("postgres"); err != nil {
		return nil, nil, fmt.Errorf("failed to set goose dialect: %v", err)
	}

	if err = goose.Up(db, "../../../../migration"); err != nil {
		return nil, nil, fmt.Errorf("failed to run goose migrations: %v", err)
	}

	cleanup := func() {
		testPool.Close()
		_ = pgContainer.Terminate(context.Background())
	}

	return testPool, cleanup, nil
}
