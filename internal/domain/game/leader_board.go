package game

import (
	"fmt"

	"github.com/google/uuid"
)

type LeaderBoard struct {
	players []PlayerWinRatio
}

func NewLeaderBoard(count int) *LeaderBoard {
	return &LeaderBoard{
		players: make([]PlayerWinRatio, 0, count),
	}
}

func (b *LeaderBoard) Add(player ...PlayerWinRatio) {
	b.players = append(b.players, player...)
}

func (b *LeaderBoard) Clone() *LeaderBoard {
	lb := NewLeaderBoard(len(b.players))
	copy(lb.players, b.players)

	return lb
}

func (b *LeaderBoard) GetPlayers() []PlayerWinRatio {
	response := b.Clone()

	return response.players
}

func (b *LeaderBoard) Length() int {
	return len(b.players)
}

type PlayerWinRatio struct {
	userID uuid.UUID
	login  string
	wins   int
	loss   int
	draw   int
	ratio  float32
}

func NewPlayerWinRatio(userID uuid.UUID, login string, wins, loss, draw int, ratio float32) (*PlayerWinRatio, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("cannot create PlayerWinRatio with nil userID")
	}

	switch {
	case wins < 0:
		wins = 0
	case loss < 0:
		loss = 0
	case draw < 0:
		draw = 0
	}

	return &PlayerWinRatio{
		userID: userID,
		login:  login,
		wins:   wins,
		loss:   loss,
		draw:   draw,
		ratio:  ratio,
	}, nil
}

func (r *PlayerWinRatio) ID() uuid.UUID {
	return r.userID
}

func (r *PlayerWinRatio) Wins() int {
	return r.wins
}

func (r *PlayerWinRatio) Loss() int {
	return r.loss
}

func (r *PlayerWinRatio) Draw() int {
	return r.draw
}

func (r *PlayerWinRatio) Ratio() float32 {
	return r.ratio
}

func (r *PlayerWinRatio) Login() string {
	return r.login
}
