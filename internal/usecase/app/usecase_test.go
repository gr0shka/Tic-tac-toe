package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/app"
)

type mockGameService struct {
	gb     *game.GameBoard
	err    error
	winner int
	status game.GameStatus
}

func (m mockGameService) GetNextTurn(ctx context.Context, gb *game.GameBoard) *game.GameBoard {
	return m.gb
}

func (m mockGameService) ValidateBoard(oldB, newB *game.GameBoard) error {
	return m.err
}

func (m mockGameService) IsEnded(board [3][3]int) (int, game.GameStatus) {
	return m.winner, m.status
}

type mockRepository struct {
	cg      *game.CurrentGame
	sliceCG []*game.CurrentGame
	err     error
}

func (m mockRepository) AllGames(ctx context.Context) ([]*game.CurrentGame, error) {
	return m.sliceCG, m.err
}

func (m mockRepository) Save(ctx context.Context, cg *game.CurrentGame) error {
	return m.err
}

func (m mockRepository) Get(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	return m.cg, m.err
}

func (m mockRepository) Update(ctx context.Context, cg *game.CurrentGame) error {
	return m.err
}

type mockTransactor struct{}

func (m mockTransactor) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}

func TestAppService_CreateGame(t *testing.T) {
	type testCase struct {
		name    string
		err     error
		repoErr error
		gameErr error
		winner  int
		status  game.GameStatus
	}

	testCases := []testCase{
		{
			name:    "success create game",
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			status:  game.StatusPlayerWins,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := game.NewGameBoard()
			gs := mockGameService{
				err:    tc.gameErr,
				winner: tc.winner,
				status: tc.status,
				gb:     gb,
			}

			cg := game.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := app.NewAppService(gs, repo, mockTransactor{})

			gotCg, err := as.CreateGameWithBot(context.Background(), uuid.New())

			if !errors.Is(err, tc.err) {
				t.Errorf("CreateGameWithBot() error = %v, wantErr %v", err, tc.err)
			}

			if tc.err == nil && gotCg == nil {
				t.Errorf("CreateGameWithBot() gotCg = <nil>, want <not nil>")
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
		isEnded game.GameStatus
	}

	testCases := []testCase{
		{
			name:    "success get game",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			isEnded: game.StatusPlayerWins,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := game.NewGameBoard()
			gs := mockGameService{
				err:    tc.gameErr,
				winner: tc.winner,
				status: tc.isEnded,
				gb:     gb,
			}

			cg := game.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := app.NewAppService(gs, repo, mockTransactor{})

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
		winner  int
		status  game.GameStatus
	}

	testCases := []testCase{
		{
			name:    "success get game",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			winner:  -1,
			status:  game.StatusWaitingForPlayers,
		},
		{
			name:    "game not found",
			id:      uuid.New(),
			err:     game.ErrNotFound,
			repoErr: game.ErrNotFound,
			winner:  0,
			status:  game.StatusDraw,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := game.NewGameBoard()
			gs := mockGameService{
				winner: tc.winner,
				status: tc.status,
				gb:     gb,
			}

			cg := game.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := app.NewAppService(gs, repo, mockTransactor{})

			winner, ended := as.GameIsEnded(context.Background(), tc.id)

			if winner != tc.winner {
				t.Errorf("GameIsEnded() winner = %v, want %v", winner, tc.winner)
			}

			if ended != tc.status {
				t.Errorf("GameIsEnded() ended = %v, want %v", ended, tc.status)
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
		status  game.GameStatus
	}

	testCases := []testCase{
		{
			name:    "success move, game not ended",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			status:  game.StatusPlayerTurn,
		},
		{
			name:    "success move, game ended, bot wins",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  1,
			status:  game.StatusPlayerWins,
		},
		{
			name:    "success move, game ended, draw",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			status:  game.StatusDraw,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := game.NewGameBoard()
			gs := mockGameService{
				err:    tc.gameErr,
				winner: tc.winner,
				status: tc.status,
				gb:     gb,
			}

			player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
			_ = gb.AddPlayer(player1)

			player2 := game.NewPlayer(uuid.New(), game.SecondPlayer, false)
			_ = gb.AddPlayer(player2)

			cg := game.NewCurrentGame(gb)
			cg.SetActivePlayer(player1)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := app.NewAppService(gs, repo, mockTransactor{})

			gotCg, err := as.ProcessPlayerMove(context.Background(), player1.ID(), tc.id, [3][3]int{})

			if !errors.Is(err, tc.err) {
				t.Errorf("ProcessPlayerMove() error = %v, wantErr %v", err, tc.err)

			}

			if tc.err == nil && gotCg == nil {
				t.Errorf("ProcessPlayerMove() gotCg = <nil>, want <not nil>")
			}

			if tc.err == nil && gotCg.Winner() != tc.winner {
				t.Errorf("ProcessPlayerMove() winner = %v, want %v", gotCg.Winner(), tc.winner)
			}

			if tc.err == nil && gotCg.Status() != tc.status {
				t.Errorf("ProcessPlayerMove() ended = %v, want %v", gotCg.IsEnded(), tc.status)
			}
		})
	}
}

func TestAppService_JoinGame(t *testing.T) {
	type testCase struct {
		name     string
		cg       *game.CurrentGame
		gb       *game.GameBoard
		gameID   uuid.UUID
		playerID uuid.UUID
		err      error
	}

	player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
	player2 := game.NewPlayer(uuid.New(), game.SecondPlayer, true)

	gbWith1Player := game.NewGameBoard()
	gbWith1Player.AddPlayer(player1)

	gbWith2Player := game.NewGameBoard()
	gbWith2Player.AddPlayer(player1)
	gbWith2Player.AddPlayer(player2)

	gameID := uuid.New()
	userID := uuid.New()

	testCases := []testCase{
		{
			name:     "success join game",
			cg:       game.NewCurrentGameWithID(gameID, gbWith1Player),
			gb:       gbWith1Player,
			gameID:   gameID,
			playerID: userID,
			err:      nil,
		},
		{
			name:     "failed join game, game is full",
			cg:       game.NewCurrentGameWithID(gameID, gbWith2Player),
			gb:       gbWith2Player,
			gameID:   gameID,
			playerID: userID,
			err:      game.ErrMaxCountOfPlayers,
		},
		{
			name:     "failed join game, player already exists",
			cg:       game.NewCurrentGameWithID(gameID, gbWith1Player),
			gb:       gbWith1Player,
			gameID:   gameID,
			playerID: gbWith1Player.Players()[game.FirstPlayer].ID(),
			err:      game.ErrPlayerAlreadyExists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepository{tc.cg, nil, nil}
			gs := mockGameService{tc.gb, nil, 0, game.StatusWaitingForPlayers}
			as := app.NewAppService(gs, repo, mockTransactor{})

			cg, err := as.JoinGame(context.Background(), tc.gameID, tc.playerID)

			if !errors.Is(err, tc.err) {
				t.Errorf("JoinGame() error = %v, wantErr %v", err, tc.err)
			}

			if err == nil {

				if len(cg.Players()) != game.CountPlayers {
					t.Errorf("JoinGame() count players got = %v, want = %v", len(cg.Players()), game.CountPlayers)
				}

				if len(cg.Players()) == game.CountPlayers {

					if cg.Players()[game.SecondPlayer].ID() != tc.playerID {
						t.Errorf("JoinGame() second player_id got = %v want = %v",
							cg.Players()[game.SecondPlayer].ID(), tc.playerID)
					}

					if cg.Players()[game.SecondPlayer].IsRealPlayer() != tc.cg.Players()[game.SecondPlayer].IsRealPlayer() {
						t.Errorf("JoinGame() second player_isRealPlayer got = %v want = %v",
							cg.Players()[game.SecondPlayer].IsRealPlayer(), tc.cg.Players()[game.SecondPlayer].IsRealPlayer())
					}

					if cg.Players()[game.SecondPlayer].Symbol() != tc.cg.Players()[game.SecondPlayer].Symbol() {
						t.Errorf("JoinGame() second player_symbol got = %v want = %v",
							cg.Players()[game.SecondPlayer].Symbol(), tc.cg.Players()[game.SecondPlayer].Symbol())
					}
				}
			}
		})
	}
}
