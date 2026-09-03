package models

import (
	"sync"

	"github.com/google/uuid"
)

type CurrentGame struct {
	id uuid.UUID
	*GameBoard
	mu sync.RWMutex
}

func NewCurrentGame(gb *GameBoard) *CurrentGame {
	return &CurrentGame{
		id:        uuid.New(),
		GameBoard: gb,
	}
}

func (c *CurrentGame) ID() uuid.UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.id
}
