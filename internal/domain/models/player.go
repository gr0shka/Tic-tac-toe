package models

type Player struct {
	turnNumber int
	realPlayer bool
}

func NewPlayer(turnNumber int, realPlayer bool) Player {
	return Player{
		turnNumber: turnNumber,
		realPlayer: realPlayer,
	}
}

func (player Player) GetTurnNumber() int {
	return player.turnNumber
}

func (player Player) IsRealPlayer() bool {
	return player.realPlayer
}
