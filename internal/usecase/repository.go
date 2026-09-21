package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type Repository interface {
	Save(ctx context.Context, cg *game.CurrentGame) error
	Get(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error)
}
