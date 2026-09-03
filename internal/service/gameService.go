package service

import "Tic-Tac-Toy/internal/models"

type GameService interface {
	GetNextTurn(gb models.GameBoard, realPlayer) int
}
