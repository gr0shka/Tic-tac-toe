package game

import (
	"github.com/google/uuid"
)

type PlayerDTO struct {
	ID         uuid.UUID `db:"id"`
	TurnNumber int       `db:"turn_of_number"`
	RealPlayer bool      `db:"real_player"`
}

type CurrentGameDTO struct {
	ID           uuid.UUID `db:"id"`
	Board        []int     `db:"board"`
	NumberOfTurn int       `db:"number_of_turn"`
	Player1ID    uuid.UUID `db:"player1_id"`
	Player1Real  bool      `db:"player1_real"`
	Player2ID    uuid.UUID `db:"player2_id"`
	Player2Real  bool      `db:"player2_real"`
	Status       string    `db:"status"`
	Winner       int       `db:"winner"`
}
