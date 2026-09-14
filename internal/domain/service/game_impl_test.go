package service_test

import (
	"testing"

	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

func createGameService(board [game.BoardSize][game.BoardSize]int, numberOfTurn int) *game.GameBoard {
	gb := game.NewGameBoard()
	gb.SetBoard(board)

	playerReal := game.NewPlayer(0, true)
	playerBot := game.NewPlayer(1, false)

	gb.AddPlayers(*playerReal, *playerBot)

	gb.SetTurnNumber(numberOfTurn)

	return gb
}

func TestGameService_GetNextTurn(t *testing.T) {
	type testCase struct {
		name string
		gb   *game.GameBoard
	}
}
