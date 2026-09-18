package postgres

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/jackc/pgx/v5"
)

type repository struct {
	db *pgx.Conn
}

func New(db *pgx.Conn) *repository {
	return &repository{
		db: db,
	}
}

func (m *repository) Save(cg *game.CurrentGame) error {
	return nil
}

func (m *repository) Get(id uuid.UUID) (*game.CurrentGame, error) {

	return nil, nil
}
