package usecase

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type Repository interface {
	Save(cg *game.CurrentGame) error
	Get(id uuid.UUID) (*game.CurrentGame, error)
}
