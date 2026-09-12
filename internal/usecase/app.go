package usecase

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type AppService interface {
	CreateGame() (*game.CurrentGame, error)
	ProcessPlayerMove(
		id uuid.UUID,
		board [game.BoardSize][game.BoardSize]int,
	) (*game.CurrentGame, error)
	GetGame(id uuid.UUID) (*game.CurrentGame, error)
	GameIsEnded(id uuid.UUID) (int, bool)
}
