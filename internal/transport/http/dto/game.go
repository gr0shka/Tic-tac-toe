package dto

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
)

type CreateGameResponse struct {
	Id    uuid.UUID                               `json:"id"`
	Board [models.BoardSize][models.BoardSize]int `json:"board"`
}

type GameBoardRequest struct {
	Board [models.BoardSize][models.BoardSize]int `json:"board"`
}

type GameBoardResponse struct {
	GameIsEnded bool                                    `json:"game_is_ended"`
	Winner      int                                     `json:"winner"`
	Board       [models.BoardSize][models.BoardSize]int `json:"board"`
}
