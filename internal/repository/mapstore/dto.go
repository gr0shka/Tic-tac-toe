package mapstore

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
)

type PlayerDTO struct {
	TurnNumber int
	RealPlayer bool
}

type CurrentGameDTO struct {
	ID           uuid.UUID
	Board        [models.BoardSize][models.BoardSize]int
	NumberOfTurn int
	Players      [models.CountPlayers]PlayerDTO
}
