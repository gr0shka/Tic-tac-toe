package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *repository {
	return &repository{
		db: db,
	}
}

func (m *repository) Save(ctx context.Context, cg *game.CurrentGame) error {

	return nil
}

func (m *repository) Get(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {

	return nil, nil
}
