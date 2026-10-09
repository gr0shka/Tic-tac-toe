package dto

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type GameBoardRequest struct {
	Board [game.BoardSize][game.BoardSize]int `json:"board"`
}

type GameBoardResponse struct {
	ID           uuid.UUID  `json:"id"`
	Status       string     `json:"status"`
	ActivePlayer *uuid.UUID `json:"active_player"`

	Player1ID     uuid.UUID `json:"player1_id"`
	Player1Symbol int       `json:"player1_symbol"`

	Player2ID     *uuid.UUID `json:"player2_id"`
	Player2Symbol *int       `json:"player2_symbol"`

	Winner   int                                 `json:"winner"`
	WinnerID *uuid.UUID                          `json:"winner_id"`
	Board    [game.BoardSize][game.BoardSize]int `json:"board"`
}

type CreateGameRequest struct {
	Mode string `json:"mode"`
}
