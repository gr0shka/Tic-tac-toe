package game

type Player struct {
	symbol     int
	realPlayer bool
}

func NewPlayer(symbol int, realPlayer bool) *Player {
	return &Player{
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
