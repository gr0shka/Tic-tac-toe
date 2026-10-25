package rules

import (
	"context"

	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type GameRules interface {
	GetNextTurn(ctx context.Context, gb *game.GameBoard) *game.GameBoard
	ValidateBoard(oldB, newB *game.GameBoard) error
	IsEnded(board [game.BoardSize][game.BoardSize]int) (int, game.GameStatus)
}
