package game_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	game2 "github.com/gr0shka/Tic-tac-toe/internal/repository/postgres/game"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),

		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("Failed to start postgres: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connection string: %v", err)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to create test pool: %v", err)
	}

	db := stdlib.OpenDBFromPool(testPool)
	defer db.Close()

	if err = goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}

	// Накатит все *.sql файлы из папки migration
	if err = goose.Up(db, "../../../../migration"); err != nil {
		log.Fatalf("failed to run goose migrations: %v", err)
	}

	exitCode := m.Run()

	testPool.Close()
	if err = pgContainer.Terminate(ctx); err != nil {
		log.Fatalf("Failed to terminate postgres container: %v", err)
	}

	os.Exit(exitCode)
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
				t.Errorf("Board error, expected %v, got %v", testCase.cg.ID(), cgDTO.Board)
			}

			if cgDTO.NumberOfTurn != testCase.cg.TurnNumber() {
				t.Errorf("turn number error, expected %v, got %v", testCase.cg.ID(), cgDTO.Board)
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
				t.Errorf("Board error, expected %v, got %v", cg.ID(), cgDTO.Board)
			}

			if cgDTO.NumberOfTurn != cg.TurnNumber() {
				t.Errorf("turn number error, expected %v, got %v", cg.ID(), cgDTO.Board)
			}
		})
	}
}
