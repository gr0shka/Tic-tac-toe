package service

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
)

type AppService interface {
	CreateGame() (*models.CurrentGame, error)
	ProcessPlayerMove(
		id uuid.UUID,
		board [models.BoardSize][models.BoardSize]int,
	) (*models.CurrentGame, error)
	GetGame(id uuid.UUID) (*models.CurrentGame, error)
	GameIsEnded(id uuid.UUID) (int, bool)
}
