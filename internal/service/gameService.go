package service

import "Tic-Tac-Toy/internal/models"

type GameService interface {
	GetNextTurn(gb *models.GameBoard) *models.GameBoard
	ValidateBoard(oldB, newB *models.GameBoard) error
	IsEnded(gb *models.GameBoard) (int, bool)
}
