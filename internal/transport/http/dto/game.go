package dto

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type GameBoardRequest struct {
	Board [game.BoardSize][game.BoardSize]int `json:"board"`
}

type GameBoardResponse struct {
	ID          uuid.UUID                           `json:"id"`
	GameIsEnded bool                                `json:"game_is_ended"`
	Winner      int                                 `json:"winner"`
	Board       [game.BoardSize][game.BoardSize]int `json:"board"`
}
