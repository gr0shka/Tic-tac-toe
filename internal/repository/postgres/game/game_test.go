package game_test

import (
	"context"
	"errors"
	"log"
	"math"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	gameRepo "github.com/gr0shka/Tic-tac-toe/internal/repository/postgres/game"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, cleanup, err := testutil.SetupTestDB(ctx)
	if err != nil {
		log.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	testPool = pool
	os.Exit(m.Run())
}

func truncateGames(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE current_game CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate users: %v", err)
	}
}

func createCurrentGame() *game.CurrentGame {
	gb := game.NewGameBoard()

	playerReal := game.NewPlayer(uuid.New(), 0, true)
	playerBot := game.NewPlayer(uuid.New(), 1, false)

	gb.AddPlayer(playerReal)
	gb.AddPlayer(playerBot)

	cg := game.NewCurrentGame(gb)

	return cg
}

func createCurrentGameDTO() *gameRepo.CurrentGameDTO {
	cg := createCurrentGame()

	players := cg.Players()

	var p2ID *uuid.UUID
	var p2Real *bool
	var p2Symbol *int
	if len(players) > 1 && players[game.SecondPlayer] != nil {
		p2ID = new(players[game.SecondPlayer].ID())
		p2Real = new(players[game.SecondPlayer].IsRealPlayer())
		p2Symbol = new(players[game.SecondPlayer].Symbol())
	}

	var activePlayerID uuid.UUID
	if cg.ActivePlayer() != nil {
		activePlayerID = cg.ActivePlayer().ID()
	}

	cgDTO := gameRepo.CurrentGameDTO{
		ID:           cg.ID(),
		Board:        gameRepo.FlattenBoard(cg.Board()),
		NumberOfTurn: cg.TurnNumber(),
		Winner:       cg.Winner(),

		Player1ID:     players[0].ID(),
		Player1Real:   players[0].IsRealPlayer(),
		Player1Symbol: players[0].Symbol(),

		Player2ID:     p2ID,
		Player2Real:   p2Real,
		Player2Symbol: p2Symbol,

		ActivePlayerID: &activePlayerID,
	}

	return &cgDTO
}

func TestRepository_Save(t *testing.T) {
	truncateGames(t)

	type useCases struct {
		name string
		cg   *game.CurrentGame
		err  error
	}

	testCases := []useCases{
		{
			name: "Success save",
			cg:   createCurrentGame(),
			err:  nil,
		},
	}

	repo := gameRepo.New(testPool)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := repo.Save(context.Background(), testCase.cg)

			if !errors.Is(err, testCase.err) {
				t.Errorf("save error, expected %v, got %v", testCase.err, err)
			}

			_, err = repo.Get(context.Background(), testCase.cg.ID())
			if err != nil {
				t.Errorf("get error, expected %v, got %v", testCase.cg.ID(), err)
			}
		})
	}
}

func TestRepository_Get(t *testing.T) {
	truncateGames(t)

	type useCases struct {
		name string
		cg   *game.CurrentGame
		err  error
	}

	testCases := []useCases{
		{
			name: "Success get",
			cg:   createCurrentGame(),
			err:  nil,
		},
		{
			name: "Not found",
			cg:   createCurrentGame(),
			err:  game.ErrNotFound,
		},
	}

	repo := gameRepo.New(testPool)

	testCase := testCases[0]
	t.Run(testCase.name, func(t *testing.T) {
		err := repo.Save(context.Background(), testCase.cg)
		if err != nil {
			t.Errorf("save error, expected %v, got %v", testCase.err, err)
		}

		_, err = repo.Get(context.Background(), testCase.cg.ID())
		if !errors.Is(err, testCase.err) {
			t.Errorf("get error, expected %v, got %v", testCase.cg.ID(), err)
		}
	})

	testCase = testCases[1]
	t.Run(testCase.name, func(t *testing.T) {

		_, err := repo.Get(context.Background(), testCase.cg.ID())
		if !errors.Is(err, testCase.err) {
			t.Errorf("get error, expected %v, got %t", testCase.cg.ID(), err)
		}
	})
}

func TestRepository_AllGames(t *testing.T) {
	truncateGames(t)

	type useCases struct {
		name         string
		cgEnded      *game.CurrentGame
		cgWaiting    *game.CurrentGame
		countOfGames int
		err          error
	}

	player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
	player2 := game.NewPlayer(uuid.New(), game.SecondPlayer, true)

	gbWith1Player := game.NewGameBoard()
	gbWith1Player.AddPlayer(player1)

	gbWith2Player := game.NewGameBoard()
	gbWith2Player.AddPlayer(player1)
	gbWith2Player.AddPlayer(player2)

	cgWaiting := game.NewCurrentGame(gbWith1Player)
	cgWaiting.SetStatus(game.StatusWaitingForPlayers)

	cgEnded := game.NewCurrentGame(gbWith2Player)
	cgEnded.SetStatus(game.StatusPlayerWins)

	repo := gameRepo.New(testPool)

	repo.Save(context.Background(), cgEnded)
	repo.Save(context.Background(), cgWaiting)

	testCases := []useCases{
		{
			name:         "Success allGames",
			cgEnded:      cgEnded,
			cgWaiting:    cgWaiting,
			countOfGames: 1,
			err:          nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cgGot, err := repo.AllGames(context.Background())

			if !errors.Is(err, testCase.err) {
				t.Errorf("get error, expected %v, got %v", testCase.err, err)
			}

			if err == nil {

				if len(cgGot) != testCase.countOfGames {
					t.Errorf("get error, expected %v, got %v", testCase.countOfGames, len(cgGot))
				}

				if len(cgGot) >= 1 && cgGot[0].ID() != testCase.cgWaiting.ID() {
					t.Errorf("get error, expected %v, got %v", testCase.cgWaiting.ID(), cgGot[0].ID())
				}
			}
		})
	}
}

func TestRepository_Update(t *testing.T) {
	truncateGames(t)

	type useCases struct {
		name  string
		cgOld *game.CurrentGame
		cgNew *game.CurrentGame
		err   error
	}

	gb := game.NewGameBoard()
	gb.AddPlayer(game.NewPlayer(uuid.New(), game.FirstPlayer, true))
	cgOld := game.NewCurrentGame(gb)
	cgOld.SetStatus(game.StatusWaitingForPlayers)

	gbNew := gb.Clone()
	gbNew.AddPlayer(game.NewPlayer(uuid.New(), game.SecondPlayer, true))
	cgNew := game.NewCurrentGameWithID(cgOld.ID(), gbNew)
	cgNew.SetActivePlayer(cgNew.Players()[game.FirstPlayer])
	cgNew.SetTurnNumber(12)

	testCases := []useCases{
		{
			name:  "Success update",
			cgNew: cgNew,
			cgOld: cgOld,
			err:   nil,
		},
	}

	repo := gameRepo.New(testPool)
	repo.Save(context.Background(), cgOld)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := repo.Update(context.Background(), testCase.cgNew)
			if err != nil {
				t.Errorf("update error, expected %v, got %v", testCase.err, err)
			}

			cgGot, err := repo.Get(context.Background(), testCase.cgNew.ID())
			if !errors.Is(err, testCase.err) {
				t.Errorf("get error, expected %v, got %v", testCase.err, err)
			}

			if err == nil {

				if cgGot.ID() != testCase.cgNew.ID() {
					t.Errorf("get error, expected %v, got %v", testCase.cgNew.ID(), cgGot.ID())
				}

				if cgGot.Status() != testCase.cgNew.Status() {
					t.Errorf("get error, expected %v, got %v", testCase.cgNew.Status(), cgGot.Status())
				}

				if cgGot.ActivePlayer().ID() != testCase.cgNew.ActivePlayer().ID() {
					t.Errorf("get error, expected %v, got %v",
						testCase.cgNew.ActivePlayer().ID(), cgGot.ActivePlayer().ID())
				}

				if cgGot.TurnNumber() != testCase.cgNew.TurnNumber() {
					t.Errorf("get error, expected %v, got %v", testCase.cgNew.TurnNumber(), cgGot.TurnNumber())
				}
			}
		})
	}
}

func TestRepository_Mapper_DomainToDTO(t *testing.T) {
	type useCases struct {
		name string
		cg   *game.CurrentGame
	}

	testCases := []useCases{
		{
			name: "Domain to dto",
			cg:   createCurrentGame(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cgDTO := gameRepo.DomainToDTO(*testCase.cg)

			if cgDTO.ID != testCase.cg.ID() {
				t.Errorf("ID error, expected %v, got %v", testCase.cg.ID(), cgDTO.ID)
			}

			if gameRepo.UnFlattenBoard(cgDTO.Board) != testCase.cg.Board() {
				t.Errorf("Board error, expected %v, got %v", testCase.cg.Board(), gameRepo.UnFlattenBoard(cgDTO.Board))
			}

			if cgDTO.NumberOfTurn != testCase.cg.TurnNumber() {
				t.Errorf("turn number error, expected %v, got %v", testCase.cg.TurnNumber(), cgDTO.NumberOfTurn)
			}
		})
	}
}

func TestRepository_Mapper_DTOtoDomain(t *testing.T) {
	type useCases struct {
		name string
		cg   *gameRepo.CurrentGameDTO
	}

	testCases := []useCases{
		{
			name: "Domain to dto",
			cg:   createCurrentGameDTO(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cgDTO := testCase.cg
			cg := gameRepo.DTOToDomain(*testCase.cg)

			if cgDTO.ID != cg.ID() {
				t.Errorf("ID error, expected %v, got %v", cg.ID(), cgDTO.ID)
			}

			if gameRepo.UnFlattenBoard(cgDTO.Board) != cg.Board() {
				t.Errorf("Board error, expected %v, got %v",
					cg.Board(), gameRepo.UnFlattenBoard(cgDTO.Board))
			}

			if cgDTO.NumberOfTurn != cg.TurnNumber() {
				t.Errorf("turn number error, expected %v, got %v",
					cg.TurnNumber(), cgDTO.NumberOfTurn)
			}
		})
	}
}

func truncateAll(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE current_game, users CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

func createTestUser(t *testing.T, id uuid.UUID, login string) {
	t.Helper()
	_, err := testPool.Exec(
		context.Background(),
		"INSERT INTO users (id, login, password) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING",
		id, login, "hash",
	)
	if err != nil {
		t.Fatalf("failed inserting test user: %v", err)
	}
}

func createFinishedGame(p1ID uuid.UUID, p2ID uuid.UUID, p2Real bool, status game.GameStatus, winnerIdx int) *game.CurrentGame {
	gb := game.NewGameBoard()
	p1 := game.NewPlayer(p1ID, game.FirstPlayer, true)
	p2 := game.NewPlayer(p2ID, game.SecondPlayer, p2Real)
	gb.AddPlayer(p1)
	gb.AddPlayer(p2)

	cg := game.NewCurrentGame(gb)
	cg.SetStatus(status)
	cg.SetWinner(winnerIdx)
	return cg
}

func TestRepository_GetLeaderBoard(t *testing.T) {
	aliceID := uuid.New()
	bobID := uuid.New()
	charlieID := uuid.New()

	setupWithGames := func(t *testing.T) {
		truncateAll(t)
		createTestUser(t, aliceID, "alice")
		createTestUser(t, bobID, "bob")
		createTestUser(t, charlieID, "charlie")

		repo := gameRepo.New(testPool)
		games := []*game.CurrentGame{
			createFinishedGame(aliceID, bobID, true, game.StatusPlayerWins, 0),         // Alice wins vs Bob
			createFinishedGame(aliceID, bobID, true, game.StatusPlayerWins, 0),         // Alice wins vs Bob
			createFinishedGame(bobID, charlieID, true, game.StatusPlayerWins, 0),       // Bob wins vs Charlie
			createFinishedGame(aliceID, charlieID, true, game.StatusDraw, -1),          // Draw between Alice and Charlie
			createFinishedGame(aliceID, bobID, true, game.StatusWaitingForPlayers, -1), // Incomplete game, should be ignored
		}
		for _, g := range games {
			if err := repo.Save(context.Background(), g); err != nil {
				t.Fatalf("failed to save game: %v", err)
			}
		}
	}

	setupEmpty := func(t *testing.T) {
		truncateAll(t)
	}

	type expectedPlayer struct {
		id        uuid.UUID
		login     string
		wins      int
		loss      int
		draw      int
		wantRatio float32
	}

	type testCase struct {
		name        string
		setup       func(t *testing.T)
		ctxFunc     func() (context.Context, context.CancelFunc)
		count       int
		wantErr     bool
		wantLength  int
		wantPlayers []expectedPlayer
	}

	defaultCtx := func() (context.Context, context.CancelFunc) {
		return context.WithCancel(context.Background())
	}

	canceledCtx := func() (context.Context, context.CancelFunc) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx, cancel
	}

	testCases := []testCase{
		{
			name:       "success get leaderboard with ranking and stats",
			setup:      setupWithGames,
			ctxFunc:    defaultCtx,
			count:      10,
			wantErr:    false,
			wantLength: 3,
			wantPlayers: []expectedPlayer{
				{
					id:        aliceID,
					login:     "alice",
					wins:      2,
					loss:      0,
					draw:      1,
					wantRatio: 2.0 / 3.0,
				},
				{
					id:        bobID,
					login:     "bob",
					wins:      1,
					loss:      2,
					draw:      0,
					wantRatio: 1.0 / 3.0,
				},
				{
					id:        charlieID,
					login:     "charlie",
					wins:      0,
					loss:      1,
					draw:      1,
					wantRatio: 0.0,
				},
			},
		},
		{
			name:       "success get leaderboard with limit",
			setup:      setupWithGames,
			ctxFunc:    defaultCtx,
			count:      1,
			wantErr:    false,
			wantLength: 1,
			wantPlayers: []expectedPlayer{
				{
					id:        aliceID,
					login:     "alice",
					wins:      2,
					loss:      0,
					draw:      1,
					wantRatio: 2.0 / 3.0,
				},
			},
		},
		{
			name:        "success empty leaderboard when no completed games",
			setup:       setupEmpty,
			ctxFunc:     defaultCtx,
			count:       10,
			wantErr:     false,
			wantLength:  0,
			wantPlayers: nil,
		},
		{
			name:        "failed get leaderboard on canceled context",
			setup:       setupWithGames,
			ctxFunc:     canceledCtx,
			count:       10,
			wantErr:     true,
			wantLength:  0,
			wantPlayers: nil,
		},
	}

	repo := gameRepo.New(testPool)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t)
			}

			ctx, cancel := tc.ctxFunc()
			defer cancel()

			lb, err := repo.GetLeaderBoard(ctx, tc.count)
			if (err != nil) != tc.wantErr {
				t.Fatalf("GetLeaderBoard() error = %v, wantErr = %v", err, tc.wantErr)
			}

			if tc.wantErr {
				return
			}

			if lb == nil {
				t.Fatal("expected non-nil LeaderBoard")
			}

			if lb.Length() != tc.wantLength {
				t.Fatalf("LeaderBoard length = %d, want %d", lb.Length(), tc.wantLength)
			}

			players := lb.GetPlayers()
			for i, wantP := range tc.wantPlayers {
				gotP := players[i]
				if gotP.ID() != wantP.id {
					t.Errorf("player[%d] ID = %v, want %v", i, gotP.ID(), wantP.id)
				}
				if gotP.Login() != wantP.login {
					t.Errorf("player[%d] Login = %s, want %s", i, gotP.Login(), wantP.login)
				}
				if gotP.Wins() != wantP.wins {
					t.Errorf("player[%d] Wins = %d, want %d", i, gotP.Wins(), wantP.wins)
				}
				if gotP.Loss() != wantP.loss {
					t.Errorf("player[%d] Loss = %d, want %d", i, gotP.Loss(), wantP.loss)
				}
				if gotP.Draw() != wantP.draw {
					t.Errorf("player[%d] Draw = %d, want %d", i, gotP.Draw(), wantP.draw)
				}
				if math.Abs(float64(gotP.Ratio()-wantP.wantRatio)) > 0.001 {
					t.Errorf("player[%d] Ratio = %f, want %f", i, gotP.Ratio(), wantP.wantRatio)
				}
			}
		})
	}
}

func TestRepository_Mapper_DTOToLeaderBoard(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	type testCase struct {
		name       string
		input      []gameRepo.PlayerWinRatioDTO
		wantLength int
	}

	testCases := []testCase{
		{
			name: "success mapping dto to leaderboard",
			input: []gameRepo.PlayerWinRatioDTO{
				{
					ID:    id1,
					Login: "alice",
					Wins:  5,
					Loss:  2,
					Draw:  1,
					Ratio: 0.625,
				},
				{
					ID:    id2,
					Login: "bob",
					Wins:  2,
					Loss:  5,
					Draw:  1,
					Ratio: 0.25,
				},
			},
			wantLength: 2,
		},
		{
			name: "skip player with nil uuid",
			input: []gameRepo.PlayerWinRatioDTO{
				{
					ID:    uuid.Nil,
					Login: "invalid",
					Wins:  0,
					Loss:  0,
					Draw:  0,
					Ratio: 0,
				},
			},
			wantLength: 0,
		},
		{
			name:       "empty dto list",
			input:      []gameRepo.PlayerWinRatioDTO{},
			wantLength: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lb := gameRepo.DTOToLeaderBoard(tc.input)
			if lb == nil {
				t.Fatal("expected non-nil LeaderBoard")
			}
			if lb.Length() != tc.wantLength {
				t.Errorf("length = %d, want %d", lb.Length(), tc.wantLength)
			}
			if len(tc.input) > 0 && tc.wantLength > 0 {
				p := lb.GetPlayers()[0]
				if p.ID() != tc.input[0].ID {
					t.Errorf("ID = %v, want %v", p.ID(), tc.input[0].ID)
				}
				if p.Login() != tc.input[0].Login {
					t.Errorf("Login = %v, want %v", p.Login(), tc.input[0].Login)
				}
				if p.Wins() != tc.input[0].Wins {
					t.Errorf("Wins = %v, want %v", p.Wins(), tc.input[0].Wins)
				}
				if p.Loss() != tc.input[0].Loss {
					t.Errorf("Loss = %v, want %v", p.Loss(), tc.input[0].Loss)
				}
				if p.Draw() != tc.input[0].Draw {
					t.Errorf("Draw = %v, want %v", p.Draw(), tc.input[0].Draw)
				}
			}
		})
	}
}
