package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/config"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/postgres"
)

func createCurrentGame() *game.CurrentGame {
	gb := game.NewGameBoard()

	playerReal := game.NewPlayer(uuid.New(), 0, true)
	playerBot := game.NewPlayer(uuid.New(), 1, false)

	gb.AddPlayers(*playerReal, *playerBot)

	cg := game.NewCurrentGame(gb)

	return cg
}

func createCurrentGameDTO() *postgres.CurrentGameDTO {
	cg := createCurrentGame()

	players := cg.Players()

	cgDTO := postgres.CurrentGameDTO{
		ID:           cg.ID(),
		Board:        postgres.FlattenBoard(cg.Board()),
		NumberOfTurn: cg.TurnNumber(),
		Player1ID:    players[0].ID(),
		Player2ID:    players[1].ID(),
		Player1Real:  players[0].IsRealPlayer(),
		Player2Real:  players[1].IsRealPlayer(),
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

	cf, err := config.Load()
	if err != nil {
		t.Skip("Could not load config file. Skipping.")
	}
	store, err := postgres.NewClient(context.Background(), cf.Postgres)
	if err != nil {
		t.Skip("Could not connect to database. Skipping.")
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repo := postgres.New(store)
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

	cf, err := config.Load()
	if err != nil {
		t.Skip(fmt.Sprintf("Could not load config file. Skipping. Error: %v", err))
	}
	store, err := postgres.NewClient(context.Background(), cf.Postgres)
	if err != nil {
		t.Skip("Could not connect to database. Skipping.")
	}

	testCase := testCases[0]
	t.Run(testCase.name, func(t *testing.T) {
		repo := postgres.New(store)
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
		repo := postgres.New(store)

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
			cgDTO := postgres.DomainToDTO(*testCase.cg)

			if cgDTO.ID != testCase.cg.ID() {
				t.Errorf("ID error, expected %v, got %v", testCase.cg.ID(), cgDTO.ID)
			}

			if postgres.UnFlattenBoard(cgDTO.Board) != testCase.cg.Board() {
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
		cg   *postgres.CurrentGameDTO
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
			cg := postgres.DTOToDomain(*testCase.cg)

			if cgDTO.ID != cg.ID() {
				t.Errorf("ID error, expected %v, got %v", cg.ID(), cgDTO.ID)
			}

			if postgres.UnFlattenBoard(cgDTO.Board) != cg.Board() {
				t.Errorf("Board error, expected %v, got %v", cg.ID(), cgDTO.Board)
			}

			if cgDTO.NumberOfTurn != cg.TurnNumber() {
				t.Errorf("turn number error, expected %v, got %v", cg.ID(), cgDTO.Board)
			}
		})
	}
}
