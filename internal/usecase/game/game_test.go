package game_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	gameRules "github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/game"
)

type mockGameRules struct {
	gb     *gameRules.GameBoard
	err    error
	winner int
	status gameRules.GameStatus
}

func (m mockGameRules) GetNextTurn(ctx context.Context, gb *gameRules.GameBoard) *gameRules.GameBoard {
	return m.gb
}

func (m mockGameRules) ValidateBoard(oldB, newB *gameRules.GameBoard) error {
	return m.err
}

func (m mockGameRules) IsEnded(board [3][3]int) (int, gameRules.GameStatus) {
	return m.winner, m.status
}

type mockRepository struct {
	cg      *gameRules.CurrentGame
	sliceCG []*gameRules.CurrentGame
	err     error
}

func (m mockRepository) GetLeaderBoard(ctx context.Context, count int) (*gameRules.LeaderBoard, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockRepository) AllEndedGames(ctx context.Context, userID uuid.UUID) ([]*gameRules.CurrentGame, error) {
	return m.sliceCG, m.err
}

func (m mockRepository) AllGames(ctx context.Context) ([]*gameRules.CurrentGame, error) {
	return m.sliceCG, m.err
}

func (m mockRepository) Save(ctx context.Context, cg *gameRules.CurrentGame) error {
	return m.err
}

func (m mockRepository) Get(ctx context.Context, id uuid.UUID) (*gameRules.CurrentGame, error) {
	return m.cg, m.err
}

func (m mockRepository) Update(ctx context.Context, cg *gameRules.CurrentGame) error {
	return m.err
}

type mockTransactor struct{}

func (m mockTransactor) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}

func TestGameService_CreateGameWithBot(t *testing.T) {
	type testCase struct {
		name    string
		err     error
		repoErr error
		gameErr error
		winner  int
		status  gameRules.GameStatus
	}

	testCases := []testCase{
		{
			name:    "success create game",
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			status:  gameRules.StatusPlayerWins,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := gameRules.NewGameBoard()
			gs := mockGameRules{
				err:    tc.gameErr,
				winner: tc.winner,
				status: tc.status,
				gb:     gb,
			}

			cg := gameRules.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := game.NewGameService(gs, repo, mockTransactor{})

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

func TestGameService_CreateGame(t *testing.T) {
	validID := uuid.New()
	saveErr := errors.New("db save error")

	type testCase struct {
		name            string
		id              uuid.UUID
		mode            gameRules.GameMode
		repoErr         error
		wantErr         error
		wantStatus      gameRules.GameStatus
		wantPlayerCount int
	}

	testCases := []testCase{
		{
			name:            "success create game player vs bot",
			id:              validID,
			mode:            gameRules.GameModePlayerVSBot,
			repoErr:         nil,
			wantErr:         nil,
			wantStatus:      gameRules.StatusPlayerTurn,
			wantPlayerCount: 2,
		},
		{
			name:            "success create game player vs player",
			id:              validID,
			mode:            gameRules.GameModePlayerVSPlayer,
			repoErr:         nil,
			wantErr:         nil,
			wantStatus:      gameRules.StatusWaitingForPlayers,
			wantPlayerCount: 1,
		},
		{
			name:    "failed nil user id",
			id:      uuid.Nil,
			mode:    gameRules.GameModePlayerVSBot,
			wantErr: gameRules.ErrInvalidArgument,
		},
		{
			name:    "failed invalid mode",
			id:      validID,
			mode:    gameRules.GameMode("INVALID_MODE"),
			wantErr: gameRules.ErrInvalidArgument,
		},
		{
			name:    "failed repo save error",
			id:      validID,
			mode:    gameRules.GameModePlayerVSPlayer,
			repoErr: saveErr,
			wantErr: saveErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepository{err: tc.repoErr}
			gs := mockGameRules{}
			as := game.NewGameService(gs, repo, mockTransactor{})

			cg, err := as.CreateGame(context.Background(), tc.id, tc.mode)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("CreateGame() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr == nil {
				if cg == nil {
					t.Fatal("CreateGame() returned nil game, want non-nil")
				}
				if cg.Status() != tc.wantStatus {
					t.Errorf("CreateGame() status = %v, wantStatus %v", cg.Status(), tc.wantStatus)
				}
				if len(cg.Players()) != tc.wantPlayerCount {
					t.Errorf("CreateGame() players count = %d, want %d", len(cg.Players()), tc.wantPlayerCount)
				}
			}
		})
	}
}

func TestGameService_GetGame(t *testing.T) {
	type testCase struct {
		name    string
		id      uuid.UUID
		err     error
		repoErr error
		gameErr error
		winner  int
		isEnded gameRules.GameStatus
	}

	testCases := []testCase{
		{
			name:    "success get game",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			isEnded: gameRules.StatusPlayerWins,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := gameRules.NewGameBoard()
			gs := mockGameRules{
				err:    tc.gameErr,
				winner: tc.winner,
				status: tc.isEnded,
				gb:     gb,
			}

			cg := gameRules.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := game.NewGameService(gs, repo, mockTransactor{})

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

func TestGameService_GameIsEnded(t *testing.T) {
	type testCase struct {
		name    string
		id      uuid.UUID
		err     error
		repoErr error
		winner  int
		status  gameRules.GameStatus
	}

	testCases := []testCase{
		{
			name:    "success get game",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			winner:  -1,
			status:  gameRules.StatusWaitingForPlayers,
		},
		{
			name:    "game not found",
			id:      uuid.New(),
			err:     gameRules.ErrNotFound,
			repoErr: gameRules.ErrNotFound,
			winner:  0,
			status:  gameRules.StatusDraw,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := gameRules.NewGameBoard()
			gs := mockGameRules{
				winner: tc.winner,
				status: tc.status,
				gb:     gb,
			}

			cg := gameRules.NewCurrentGame(gb)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := game.NewGameService(gs, repo, mockTransactor{})

			winner, ended, err := as.GameIsEnded(context.Background(), tc.id)

			if !errors.Is(err, tc.err) {
				t.Errorf("GameIsEnded() error = %v, wantErr %v", err, tc.err)
			}

			if winner != tc.winner {
				t.Errorf("GameIsEnded() winner = %v, want %v", winner, tc.winner)
			}

			if ended != tc.status {
				t.Errorf("GameIsEnded() ended = %v, want %v", ended, tc.status)
			}
		})
	}
}

func TestGameService_ProcessPlayerMove(t *testing.T) {
	type testCase struct {
		name    string
		id      uuid.UUID
		err     error
		repoErr error
		gameErr error
		winner  int
		status  gameRules.GameStatus
	}

	testCases := []testCase{
		{
			name:    "success move, game not ended",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			status:  gameRules.StatusPlayerTurn,
		},
		{
			name:    "success move, game ended, bot wins",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  1,
			status:  gameRules.StatusPlayerWins,
		},
		{
			name:    "success move, game ended, draw",
			id:      uuid.New(),
			err:     nil,
			repoErr: nil,
			gameErr: nil,
			winner:  -1,
			status:  gameRules.StatusDraw,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gb := gameRules.NewGameBoard()
			gs := mockGameRules{
				err:    tc.gameErr,
				winner: tc.winner,
				status: tc.status,
				gb:     gb,
			}

			player1 := gameRules.NewPlayer(uuid.New(), gameRules.FirstPlayer, true)
			_ = gb.AddPlayer(player1)

			player2 := gameRules.NewPlayer(uuid.New(), gameRules.SecondPlayer, false)
			_ = gb.AddPlayer(player2)

			cg := gameRules.NewCurrentGame(gb)
			cg.SetActivePlayer(player1)
			repo := mockRepository{
				cg:  cg,
				err: tc.repoErr,
			}

			as := game.NewGameService(gs, repo, mockTransactor{})

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

func TestGameService_JoinGame(t *testing.T) {
	type testCase struct {
		name     string
		cg       *gameRules.CurrentGame
		gb       *gameRules.GameBoard
		gameID   uuid.UUID
		playerID uuid.UUID
		err      error
	}

	player1 := gameRules.NewPlayer(uuid.New(), gameRules.FirstPlayer, true)
	player2 := gameRules.NewPlayer(uuid.New(), gameRules.SecondPlayer, true)

	gbWith1Player := gameRules.NewGameBoard()
	gbWith1Player.AddPlayer(player1)

	gbWith2Player := gameRules.NewGameBoard()
	gbWith2Player.AddPlayer(player1)
	gbWith2Player.AddPlayer(player2)

	gameID := uuid.New()
	userID := uuid.New()

	testCases := []testCase{
		{
			name:     "success join game",
			cg:       gameRules.NewCurrentGameWithID(gameID, gbWith1Player),
			gb:       gbWith1Player,
			gameID:   gameID,
			playerID: userID,
			err:      nil,
		},
		{
			name:     "failed join game, game is full",
			cg:       gameRules.NewCurrentGameWithID(gameID, gbWith2Player),
			gb:       gbWith2Player,
			gameID:   gameID,
			playerID: userID,
			err:      gameRules.ErrGameFull,
		},
		{
			name:     "failed join game, player already exists",
			cg:       gameRules.NewCurrentGameWithID(gameID, gbWith1Player),
			gb:       gbWith1Player,
			gameID:   gameID,
			playerID: gbWith1Player.Players()[gameRules.FirstPlayer].ID(),
			err:      gameRules.ErrPlayerAlreadyInGame,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepository{tc.cg, nil, nil}
			gs := mockGameRules{tc.gb, nil, 0, gameRules.StatusWaitingForPlayers}
			as := game.NewGameService(gs, repo, mockTransactor{})

			cg, err := as.JoinGame(context.Background(), tc.gameID, tc.playerID)

			if !errors.Is(err, tc.err) {
				t.Errorf("JoinGame() error = %v, wantErr %v", err, tc.err)
			}

			if err == nil {

				if len(cg.Players()) != gameRules.CountPlayers {
					t.Errorf("JoinGame() count players got = %v, want = %v", len(cg.Players()), gameRules.CountPlayers)
				}

				if len(cg.Players()) == gameRules.CountPlayers {

					if cg.Players()[gameRules.SecondPlayer].ID() != tc.playerID {
						t.Errorf("JoinGame() second player_id got = %v want = %v",
							cg.Players()[gameRules.SecondPlayer].ID(), tc.playerID)
					}

					if cg.Players()[gameRules.SecondPlayer].IsRealPlayer() != tc.cg.Players()[gameRules.SecondPlayer].IsRealPlayer() {
						t.Errorf("JoinGame() second player_isRealPlayer got = %v want = %v",
							cg.Players()[gameRules.SecondPlayer].IsRealPlayer(), tc.cg.Players()[gameRules.SecondPlayer].IsRealPlayer())
					}

					if cg.Players()[gameRules.SecondPlayer].Symbol() != tc.cg.Players()[gameRules.SecondPlayer].Symbol() {
						t.Errorf("JoinGame() second player_symbol got = %v want = %v",
							cg.Players()[gameRules.SecondPlayer].Symbol(), tc.cg.Players()[gameRules.SecondPlayer].Symbol())
					}
				}
			}
		})
	}
}

func TestGameService_AllGames(t *testing.T) {
	repoErr := errors.New("db error")

	cg1 := gameRules.NewCurrentGame(gameRules.NewGameBoard())
	cg2 := gameRules.NewCurrentGame(gameRules.NewGameBoard())

	type testCase struct {
		name      string
		mockGames []*gameRules.CurrentGame
		repoErr   error
		wantErr   error
		wantCount int
	}

	testCases := []testCase{
		{
			name:      "success get all games",
			mockGames: []*gameRules.CurrentGame{cg1, cg2},
			repoErr:   nil,
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:      "success empty games list",
			mockGames: []*gameRules.CurrentGame{},
			repoErr:   nil,
			wantErr:   nil,
			wantCount: 0,
		},
		{
			name:      "failed repo error",
			mockGames: nil,
			repoErr:   repoErr,
			wantErr:   repoErr,
			wantCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepository{
				sliceCG: tc.mockGames,
				err:     tc.repoErr,
			}
			gs := mockGameRules{}
			as := game.NewGameService(gs, repo, mockTransactor{})

			games, err := as.AllGames(context.Background())

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("AllGames() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr == nil {
				if len(games) != tc.wantCount {
					t.Errorf("AllGames() count = %d, wantCount %d", len(games), tc.wantCount)
				}
			}
		})
	}
}

func TestGameService_AllEndedGamesByUserID(t *testing.T) {
	userID := uuid.New()
	repoErr := errors.New("db error")

	cg1 := gameRules.NewCurrentGame(gameRules.NewGameBoard())
	cg1.SetStatus(gameRules.StatusPlayerWins)

	cg2 := gameRules.NewCurrentGame(gameRules.NewGameBoard())
	cg2.SetStatus(gameRules.StatusDraw)

	type testCase struct {
		name      string
		userID    uuid.UUID
		mockGames []*gameRules.CurrentGame
		repoErr   error
		wantErr   error
		wantCount int
	}

	testCases := []testCase{
		{
			name:      "success get ended games",
			userID:    userID,
			mockGames: []*gameRules.CurrentGame{cg1, cg2},
			repoErr:   nil,
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:      "success empty ended games",
			userID:    userID,
			mockGames: []*gameRules.CurrentGame{},
			repoErr:   nil,
			wantErr:   nil,
			wantCount: 0,
		},
		{
			name:      "failed repo error",
			userID:    userID,
			mockGames: nil,
			repoErr:   repoErr,
			wantErr:   repoErr,
			wantCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepository{
				sliceCG: tc.mockGames,
				err:     tc.repoErr,
			}
			gs := mockGameRules{}
			as := game.NewGameService(gs, repo, mockTransactor{})

			games, err := as.AllEndedGamesByUserID(context.Background(), tc.userID)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("AllEndedGamesByUserID() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr == nil {
				if len(games) != tc.wantCount {
					t.Errorf("AllEndedGamesByUserID() count = %d, wantCount %d", len(games), tc.wantCount)
				}
			}
		})
	}
}
