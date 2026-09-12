package service

import (
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type GameService interface {
	GetNextTurn(gb *game.GameBoard) *game.GameBoard
	ValidateBoard(oldB, newB *game.GameBoard) error
	IsEnded(board [game.BoardSize][game.BoardSize]int) (int, bool)
}
