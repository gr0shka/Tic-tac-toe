package game

import "github.com/google/uuid"

type Player struct {
	id         uuid.UUID
	symbol     int
	realPlayer bool
}

func NewPlayer(id uuid.UUID, symbol int, realPlayer bool) *Player {
	return &Player{
		id:         id,
		symbol:     symbol,
		realPlayer: realPlayer,
	}
}

func (player Player) Symbol() int {
	return player.symbol
}

func (player Player) IsRealPlayer() bool {
	return player.realPlayer
}

func (player Player) ID() uuid.UUID {
	return player.id
}
