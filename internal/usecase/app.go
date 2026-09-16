package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type AppService interface {
	CreateGame(ctx context.Context) (*game.CurrentGame, error)
	ProcessPlayerMove(
		ctx context.Context,
		id uuid.UUID,
		board [game.BoardSize][game.BoardSize]int,
	) (*game.CurrentGame, error)
	GetGame(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error)
	GameIsEnded(ctx context.Context, id uuid.UUID) (int, bool)
}
