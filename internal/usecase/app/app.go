package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type AppService interface {
	CreateGame(ctx context.Context, id uuid.UUID, mode game.GameMode) (*game.CurrentGame, error)
	ProcessPlayerMove(
		ctx context.Context,
		playerID, gameID uuid.UUID,
		board [game.BoardSize][game.BoardSize]int,
	) (*game.CurrentGame, error)
	GetGame(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error)
	GameIsEnded(ctx context.Context, id uuid.UUID) (int, game.GameStatus)
	JoinGame(ctx context.Context, gameID, playerID uuid.UUID) (*game.CurrentGame, error)
	AllGames(ctx context.Context) ([]*game.CurrentGame, error)
}
