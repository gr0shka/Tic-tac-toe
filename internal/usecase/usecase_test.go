package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase"
)

type mockGameService struct {
	gb      *game.GameBoard
	err     error
	winner  int
	isEnded bool
}

func (m mockGameService) GetNextTurn(ctx context.Context, gb *game.GameBoard) *game.GameBoard {
	return m.gb
}

func (m mockGameService) ValidateBoard(oldB, newB *game.GameBoard) error {
	return m.err
}

func (m mockGameService) IsEnded(board [3][3]int) (int, bool) {
	return m.winner, m.isEnded
}

type mockRepository struct {
	cg  *game.CurrentGame
	err error
}

func (m mockRepository) Save(ctx context.Context, cg *game.CurrentGame) error {
	return m.err
}

func (m mockRepository) Get(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	return m.cg, m.err
}

func TestAppService_CreateGame(t *testing.T) {
	type testCase struct {
		name    string
		err     error
		repoErr error
		gameErr error
		winner  int
		isEnded bool
	}

	testCases := []testCase{
		{
			name:    "success create game",
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			isEnded: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := game.NewGameBoard()
			gs := mockGameService{
				err:     tc.gameErr,
				winner:  tc.winner,
				isEnded: tc.isEnded,
				gb:      gb,
			}

			cg := game.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := usecase.NewAppService(gs, repo)

			gotCg, err := as.CreateGame(context.Background())

			if !errors.Is(err, tc.err) {
				t.Errorf("CreateGame() error = %v, wantErr %v", err, tc.err)
			}

			if tc.err == nil && gotCg == nil {
				t.Errorf("CreateGame() gotCg = <nil>, want <not nil>")
			}
		})
	}
}

func TestAppService_GetGame(t *testing.T) {
	type testCase struct {
		name    string
		id      uuid.UUID
		err     error
		repoErr error
		gameErr error
		winner  int
		isEnded bool
	}

	testCases := []testCase{
		{
			name:    "success get game",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			isEnded: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := game.NewGameBoard()
			gs := mockGameService{
				err:     tc.gameErr,
				winner:  tc.winner,
				isEnded: tc.isEnded,
				gb:      gb,
			}

			cg := game.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := usecase.NewAppService(gs, repo)

			gotCg, err := as.GetGame(context.Background(), tc.id)

			if !errors.Is(err, tc.err) {
				t.Errorf("GetGame() error = %v, wantErr %v", err, tc.err)
			}

			if tc.err == nil && gotCg == nil {
				t.Errorf("GetGame() gotCg = <nil>, want <not nil>")
			}
		})
	}
}

func TestAppService_GameIsEnded(t *testing.T) {
	type testCase struct {
		name    string
		id      uuid.UUID
		err     error
		repoErr error
		gameErr error
		winner  int
		isEnded bool
	}

	testCases := []testCase{
		{
			name:    "success get game",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			isEnded: false,
		},
		{
			name:    "game not found",
			id:      uuid.New(),
			err:     game.ErrNotFound,
			repoErr: game.ErrNotFound,
			gameErr: nil,
			winner:  0,
			isEnded: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := game.NewGameBoard()
			gs := mockGameService{
				err:     tc.gameErr,
				winner:  tc.winner,
				isEnded: tc.isEnded,
				gb:      gb,
			}

			cg := game.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := usecase.NewAppService(gs, repo)

			winner, ended := as.GameIsEnded(context.Background(), tc.id)

			if winner != tc.winner {
				t.Errorf("GameIsEnded() winner = %v, want %v", winner, tc.winner)
			}

			if ended != tc.isEnded {
				t.Errorf("GameIsEnded() ended = %v, want %v", ended, tc.isEnded)
			}
		})
	}
}

func TestAppService_ProcessPlayerMove(t *testing.T) {
	type testCase struct {
		name    string
		id      uuid.UUID
		err     error
		repoErr error
		gameErr error
		winner  int
		isEnded bool
	}

	testCases := []testCase{
		{
			name:    "success move, game not ended",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			isEnded: false,
		},
		{
			name:    "success move, game ended, bot wins",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  1,
			isEnded: true,
		},
		{
			name:    "success move, game ended, draw",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			isEnded: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := game.NewGameBoard()
			gs := mockGameService{
				err:     tc.gameErr,
				winner:  tc.winner,
				isEnded: tc.isEnded,
				gb:      gb,
			}

			cg := game.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := usecase.NewAppService(gs, repo)

			gotCg, err := as.ProcessPlayerMove(context.Background(), tc.id, [3][3]int{})

			if !errors.Is(err, tc.err) {
				t.Errorf("GameIsEnded() error = %v, wantErr %v", err, tc.err)

			}

			if tc.err == nil && gotCg == nil {
				t.Errorf("GameIsEnded() gotCg = <nil>, want <not nil>")
			}

			if gotCg.Winner() != tc.winner {
				t.Errorf("GameIsEnded() winner = %v, want %v", gotCg.Winner(), tc.winner)
			}

			if gotCg.IsEnded() != tc.isEnded {
				t.Errorf("GameIsEnded() ended = %v, want %v", gotCg.IsEnded(), tc.isEnded)
			}
		})
	}
}
