package mapstore_test

import (
	"errors"
	"testing"

	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/mapstore"
)

func createCurrentGame() *game.CurrentGame {
	gb := game.NewGameBoard()

	playerReal := game.NewPlayer(0, true)
	playerBot := game.NewPlayer(1, false)

	gb.AddPlayers(*playerReal, *playerBot)

	cg := game.NewCurrentGame(gb)

	return cg
}

func createCurrentGameDTO() *mapstore.CurrentGameDTO {
	cg := createCurrentGame()

	players := cg.Players()

	pl1 := mapstore.PlayerDTO{players[0].Symbol(), players[0].IsRealPlayer()}
	pl2 := mapstore.PlayerDTO{players[1].Symbol(), players[1].IsRealPlayer()}

	cgDTO := mapstore.CurrentGameDTO{
		ID:           cg.ID(),
		Board:        cg.Board(),
		NumberOfTurn: cg.TurnNumber(),
		Players:      [2]mapstore.PlayerDTO{pl1, pl2},
	}

	return &cgDTO
}

func TestMapRepository_Save(t *testing.T) {
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

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := mapstore.NewStorage()
			repo := mapstore.NewMapRepository(store)
			err := repo.Save(testCase.cg)

			if !errors.Is(err, testCase.err) {
				t.Errorf("save error, expected %v, got %v", testCase.err, err)
			}

			_, err = repo.Get(testCase.cg.ID())
			if err != nil {
				t.Errorf("get error, expected %v, got %v", testCase.cg.ID(), err)
			}
		})
	}
}

func TestMapRepository_Get(t *testing.T) {
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

	testCase := testCases[0]
	t.Run(testCase.name, func(t *testing.T) {
		store := mapstore.NewStorage()
		repo := mapstore.NewMapRepository(store)
		err := repo.Save(testCase.cg)
		if err != nil {
			t.Errorf("save error, expected %v, got %v", testCase.err, err)
		}

		_, err = repo.Get(testCase.cg.ID())
		if !errors.Is(err, testCase.err) {
			t.Errorf("get error, expected %v, got %v", testCase.cg.ID(), err)
		}
	})

	testCase = testCases[1]
	t.Run(testCase.name, func(t *testing.T) {
		store := mapstore.NewStorage()
		repo := mapstore.NewMapRepository(store)

		_, err := repo.Get(testCase.cg.ID())
		if !errors.Is(err, testCase.err) {
			t.Errorf("get error, expected %v, got %v", testCase.cg.ID(), err)
		}
	})
}

func TestMapRepository_Mapper_DomainToDTO(t *testing.T) {
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
			cgDTO := mapstore.DomainToDTO(*testCase.cg)

			if cgDTO.ID != testCase.cg.ID() {
				t.Errorf("ID error, expected %v, got %v", testCase.cg.ID(), cgDTO.ID)
			}

			if cgDTO.Board != testCase.cg.Board() {
				t.Errorf("Board error, expected %v, got %v", testCase.cg.ID(), cgDTO.Board)
			}

			if cgDTO.NumberOfTurn != testCase.cg.TurnNumber() {
				t.Errorf("turn number error, expected %v, got %v", testCase.cg.ID(), cgDTO.Board)
			}

			for k, pl := range testCase.cg.Players() {
				if pl.Symbol() != cgDTO.Players[k].TurnNumber {
					t.Errorf("player turn number error, expected %v, got %v", testCase.cg.ID(), pl.Symbol())
				}
				if pl.IsRealPlayer() != cgDTO.Players[k].RealPlayer {
					t.Errorf("player is real player error, expected %v, got %v", testCase.cg.ID(), pl.Symbol())
				}
			}
		})
	}
}

func TestMapRepository_Mapper_DTOtoDomain(t *testing.T) {
	type useCases struct {
		name string
		cg   *mapstore.CurrentGameDTO
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
			cg := mapstore.DTOToDomain(*testCase.cg)

			if cgDTO.ID != cg.ID() {
				t.Errorf("ID error, expected %v, got %v", cg.ID(), cgDTO.ID)
			}

			if cgDTO.Board != cg.Board() {
				t.Errorf("Board error, expected %v, got %v", cg.ID(), cgDTO.Board)
			}

			if cgDTO.NumberOfTurn != cg.TurnNumber() {
				t.Errorf("turn number error, expected %v, got %v", cg.ID(), cgDTO.Board)
			}

			for k, pl := range cg.Players() {
				if pl.Symbol() != cgDTO.Players[k].TurnNumber {
					t.Errorf("player turn number error, expected %v, got %v", cg.ID(), pl.Symbol())
				}
				if pl.IsRealPlayer() != cgDTO.Players[k].RealPlayer {
					t.Errorf("player is real player error, expected %v, got %v", cg.ID(), pl.Symbol())
				}
			}
		})
	}
}
