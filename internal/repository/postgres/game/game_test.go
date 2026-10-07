package game_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	game2 "github.com/gr0shka/Tic-tac-toe/internal/repository/postgres/game"
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

func createCurrentGameDTO() *game2.CurrentGameDTO {
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

	cgDTO := game2.CurrentGameDTO{
		ID:           cg.ID(),
		Board:        game2.FlattenBoard(cg.Board()),
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

	repo := game2.New(testPool)

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

	repo := game2.New(testPool)

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

	repo := game2.New(testPool)

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

	repo := game2.New(testPool)
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
			cgDTO := game2.DomainToDTO(*testCase.cg)

			if cgDTO.ID != testCase.cg.ID() {
				t.Errorf("ID error, expected %v, got %v", testCase.cg.ID(), cgDTO.ID)
			}

			if game2.UnFlattenBoard(cgDTO.Board) != testCase.cg.Board() {
				t.Errorf("Board error, expected %v, got %v", testCase.cg.Board(), game2.UnFlattenBoard(cgDTO.Board))
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
		cg   *game2.CurrentGameDTO
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
			cg := game2.DTOToDomain(*testCase.cg)

			if cgDTO.ID != cg.ID() {
				t.Errorf("ID error, expected %v, got %v", cg.ID(), cgDTO.ID)
			}

			if game2.UnFlattenBoard(cgDTO.Board) != cg.Board() {
				t.Errorf("Board error, expected %v, got %v",
					cg.Board(), game2.UnFlattenBoard(cgDTO.Board))
			}

			if cgDTO.NumberOfTurn != cg.TurnNumber() {
				t.Errorf("turn number error, expected %v, got %v",
					cg.TurnNumber(), cgDTO.NumberOfTurn)
			}
		})
	}
}
