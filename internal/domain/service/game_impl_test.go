package service_test

import (
	"errors"
	"testing"

	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	gameService "github.com/gr0shka/Tic-tac-toe/internal/domain/service"
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

		expectedBoard      [game.BoardSize][game.BoardSize]int
		expectedTurnNumber int
	}

	testCases := []testCase{
		{
			name: "first turn",
			gb: createGameService(
				[3][3]int{
					{-1, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				0),
			expectedBoard: [game.BoardSize][game.BoardSize]int{
				{0, -1, -1},
				{-1, -1, -1},
				{-1, -1, -1},
			},
			expectedTurnNumber: 1,
		},
		{
			name: "winner turn",
			gb: createGameService(
				[3][3]int{
					{-1, 1, 0},
					{-1, 1, -1},
					{0, -1, 0}},
				5),
			expectedBoard: [game.BoardSize][game.BoardSize]int{
				{-1, 1, 0},
				{-1, 1, -1},
				{0, 1, 0},
			},
			expectedTurnNumber: 6,
		},
		{
			name: "block player winner turn",
			gb: createGameService(
				[3][3]int{
					{0, 0, -1},
					{-1, -1, -1},
					{-1, -1, 1}},
				3),
			expectedBoard: [game.BoardSize][game.BoardSize]int{
				{0, 0, 1},
				{-1, -1, -1},
				{-1, -1, 1},
			},
			expectedTurnNumber: 4,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			gs := gameService.NewGameService()
			gb := gs.GetNextTurn(ts.gb)

			if gb.TurnNumber() != ts.expectedTurnNumber {
				t.Errorf("Expected turn number %d, got %d", ts.expectedTurnNumber, gb.TurnNumber())
			}

			if gb.Board() != ts.expectedBoard {
				t.Errorf("Expected board %d, got %d", ts.expectedBoard, gb.Board())
			}
		})
	}
}

func TestGameService_ValidateBoard(t *testing.T) {
	type testCase struct {
		name string
		gb1  *game.GameBoard
		gb2  *game.GameBoard
		err  error
	}

	testCases := []testCase{
		{
			name: "valid turn",
			gb1: createGameService(
				[3][3]int{
					{-1, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				0),
			gb2: createGameService(
				[3][3]int{
					{0, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				1),
			err: nil,
		},
		{
			name: "no move",
			gb1: createGameService(
				[3][3]int{
					{-1, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				0),
			gb2: createGameService(
				[3][3]int{
					{-1, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				1),
			err: game.ErrNoMoveMade,
		},
		{
			name: "multiple move",
			gb1: createGameService(
				[3][3]int{
					{-1, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				0),
			gb2: createGameService(
				[3][3]int{
					{0, 0, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				1),
			err: game.ErrMultipleMoves,
		},
		{
			name: "wrong number of turns",
			gb1: createGameService(
				[3][3]int{
					{-1, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				0),
			gb2: createGameService(
				[3][3]int{
					{0, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				123),
			err: game.ErrWrongNumberOfTurns,
		},
		{
			name: "no move",
			gb1: createGameService(
				[3][3]int{
					{0, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				1),
			gb2: createGameService(
				[3][3]int{
					{1, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				2),
			err: game.ErrAlteredPreviousMoves,
		},
		{
			name: "wrong player move",
			gb1: createGameService(
				[3][3]int{
					{-1, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				0),
			gb2: createGameService(
				[3][3]int{
					{1, -1, -1},
					{-1, -1, -1},
					{-1, -1, -1}},
				1),
			err: game.ErrWrongPlayerMove,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			gs := gameService.NewGameService()

			gotErr := gs.ValidateBoard(ts.gb1, ts.gb2)
			if !errors.Is(ts.err, gotErr) {
				t.Errorf("Expected error: %v, got: %v", ts.err, gotErr)
			}
		})
	}
}

func TestGameService_IsEnded(t *testing.T) {
	type testCase struct {
		name   string
		board  [game.BoardSize][game.BoardSize]int
		winner int
		ended  bool
	}

	testCases := []testCase{
		{
			name: "game is not ended",
			board: [3][3]int{
				{-1, -1, -1},
				{-1, -1, -1},
				{-1, -1, -1}},
			winner: gameService.Draw,
			ended:  false,
		},
		{
			name: "game is ended with draw",
			board: [3][3]int{
				{1, 1, 0},
				{0, 0, 1},
				{1, 1, 0}},
			winner: gameService.Draw,
			ended:  true,
		},
		{
			name: "game is ended with 1 player win",
			board: [3][3]int{
				{0, 1, 0},
				{0, 0, 1},
				{1, 1, 0}},
			winner: game.FirstPlayer,
			ended:  true,
		},
		{
			name: "game is ended with 1 player win",
			board: [3][3]int{
				{1, 0, 1},
				{1, 1, 0},
				{0, 0, 1}},
			winner: game.SecondPlayer,
			ended:  true,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			gs := gameService.NewGameService()

			gotWinner, gotBool := gs.IsEnded(ts.board)
			if gotWinner != ts.winner {
				t.Errorf("Expected winner %d, got %d", ts.winner, gotWinner)
			}
			if gotBool != ts.ended {
				t.Errorf("Expected ended %t, got %t", ts.ended, gotBool)
			}
		})
	}
}
