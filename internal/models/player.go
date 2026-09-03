package models

import "sync"

type Player struct {
	turnNumber int
	realPlayer bool

	mu sync.RWMutex
}

func NewPlayer(turnNumber int, realPlayer bool) *Player {
	return &Player{
		turnNumber: turnNumber,
		realPlayer: realPlayer,
	}
}

func (player *Player) GetTurnNumber() int {
	player.mu.RLock()
	defer player.mu.RUnlock()

	return player.turnNumber
}

func (player *Player) IsRealPlayer() bool {
	player.mu.RLock()
	defer player.mu.RUnlock()

	return player.realPlayer
}
