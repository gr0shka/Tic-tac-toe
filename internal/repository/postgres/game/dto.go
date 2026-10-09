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
	ID uuid.UUID `db:"id"`

	Player1ID     uuid.UUID `db:"player1_id"`
	Player1Real   bool      `db:"player1_real"`
	Player1Symbol int       `db:"player1_symbol"`

	Player2ID     *uuid.UUID `db:"player2_id"`
	Player2Real   *bool      `db:"player2_real"`
	Player2Symbol *int       `db:"player2_symbol"`

	ActivePlayerID *uuid.UUID `db:"active_player_id"`

	Board        []int  `db:"board"`
	NumberOfTurn int    `db:"number_of_turn"`
	Status       string `db:"status"`
	Winner       int    `db:"winner"`

	WinnerID *uuid.UUID `db:"winner_id"`
}
