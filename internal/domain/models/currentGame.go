package models

import (
	"github.com/google/uuid"
)

type CurrentGame struct {
	id uuid.UUID
	*GameBoard
}

func NewCurrentGame(gb *GameBoard) *CurrentGame {
	return &CurrentGame{
		id:        uuid.New(),
		GameBoard: gb,
	}
}

func (c *CurrentGame) ID() uuid.UUID {
	return c.id
}
