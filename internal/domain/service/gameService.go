package service

import (
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
)

type GameService interface {
	GetNextTurn(gb *models.GameBoard) *models.GameBoard
	ValidateBoard(oldB, newB *models.GameBoard) error
	IsEnded(board [models.BoardSize][models.BoardSize]int) (int, bool)
}
