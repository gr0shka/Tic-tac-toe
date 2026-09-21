package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type GameRepository interface {
	Save(ctx context.Context, cg *game.CurrentGame) error
	Get(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error)
	Update(ctx context.Context, cg *game.CurrentGame) error
}
