package mapstore

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type PlayerDTO struct {
	TurnNumber int
	RealPlayer bool
}

type CurrentGameDTO struct {
	ID           uuid.UUID
	Board        [game.BoardSize][game.BoardSize]int
	NumberOfTurn int
	Players      [game.CountPlayers]PlayerDTO
}
