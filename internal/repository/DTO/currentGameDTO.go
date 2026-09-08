package DTO

import "github.com/google/uuid"

const (
	boardSize    = 3
	countPlayers = 2
)

type CurrentGameDTO struct {
	ID           uuid.UUID
	Board        [boardSize][boardSize]int
	NumberOfTurn int
	Players      [countPlayers]PlayerDTO
}
