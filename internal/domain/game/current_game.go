package game

import (
	"github.com/google/uuid"
)

type CurrentGame struct {
	id uuid.UUID
	*GameBoard
	isEnded bool
	winner  int
}

func NewCurrentGame(gb *GameBoard) *CurrentGame {
	return &CurrentGame{
		id:        uuid.New(),
		GameBoard: gb,
		isEnded:   false,
		winner:    -1,
	}
}

func NewCurrentGameWithID(id uuid.UUID, gb *GameBoard) *CurrentGame {
	return &CurrentGame{
		id:        id,
		GameBoard: gb,
		isEnded:   false,
		winner:    -1,
	}
}

func (c CurrentGame) ID() uuid.UUID {
	return c.id
}

func (c CurrentGame) IsEnded() bool {
	return c.isEnded
}

func (c CurrentGame) Winner() int {
	return c.winner
}

func (c *CurrentGame) SetIsEnded(isEnded bool) {
	c.isEnded = isEnded
}

func (c *CurrentGame) SetWinner(winner int) {
	c.winner = winner
}
