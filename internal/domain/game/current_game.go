package game

import (
	"github.com/google/uuid"
)

type GameStatus string

const (
	StatusWaitingForPlayers GameStatus = "waiting_for_players"
	StatusPlayerTurn        GameStatus = "player_turn"
	StatusDraw              GameStatus = "draw"
	StatusPlayerWins        GameStatus = "player_wins"
)

type GameMode string

const (
	GameModePlayerVSPlayer GameMode = "player_vs_player"
	GameModePlayerVSBot    GameMode = "player_vs_bot"
)

type CurrentGame struct {
	id uuid.UUID
	*GameBoard
	status       GameStatus
	activePlayer *Player
	winner       int
}

func NewCurrentGame(gb *GameBoard) *CurrentGame {
	return &CurrentGame{
		id:           uuid.New(),
		GameBoard:    gb,
		status:       StatusWaitingForPlayers,
		winner:       -1,
		activePlayer: nil,
	}
}

func NewCurrentGameWithID(id uuid.UUID, gb *GameBoard) *CurrentGame {
	return &CurrentGame{
		id:           id,
		GameBoard:    gb,
		status:       StatusWaitingForPlayers,
		winner:       -1,
		activePlayer: nil,
	}
}

func (c CurrentGame) ID() uuid.UUID {
	return c.id
}

func (c CurrentGame) IsEnded() bool {
	return c.status == StatusDraw || c.status == StatusPlayerWins
}

func (c CurrentGame) Winner() int {
	return c.winner
}

func (c CurrentGame) Status() GameStatus {
	return c.status
}

func (c *CurrentGame) SetStatus(gs GameStatus) {
	c.status = gs
}

func (c *CurrentGame) SetWinner(winner int) {
	c.winner = winner
}

func (c *CurrentGame) SetActivePlayer(p *Player) {
	c.activePlayer = p
}

func (c *CurrentGame) ActivePlayer() *Player {
	return c.activePlayer
}
