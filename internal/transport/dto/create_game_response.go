package dto

import "github.com/google/uuid"

type CreateGameResponse struct {
	Id    uuid.UUID
	Board [boardSize][boardSize]int `json:"board"`
}
