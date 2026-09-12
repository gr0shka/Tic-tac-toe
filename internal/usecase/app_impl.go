package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/service"
	"github.com/gr0shka/Tic-tac-toe/internal/repository"
)

type appService struct {
	gameService service.GameService
	repository  repository.Repository
}

func (a appService) GameIsEnded(id uuid.UUID) (int, bool) {
	cg, err := a.repository.Get(id)
	if err != nil {
		return 0, false
	}

	return a.gameService.IsEnded(cg.Board())
}

func NewAppService(gameService service.GameService, rep repository.Repository) *appService {
	return &appService{
		gameService: gameService,
		repository:  rep,
	}
}

func (a appService) CreateGame() (*models.CurrentGame, error) {
	gb := models.NewGameBoard()
	player := models.NewPlayer(models.FirstPlayer, true)
	computer := models.NewPlayer(models.SecondPlayer, false)
	gb.AddPlayers(*player, *computer)

	cg := models.NewCurrentGame(gb)

	if err := a.repository.Save(*cg); err != nil {
		return nil, err
	}

	return cg, nil
}

func (a appService) ProcessPlayerMove(
	id uuid.UUID,
	board [models.BoardSize][models.BoardSize]int,
) (*models.CurrentGame, error) {

	current, err := a.repository.Get(id)
	if err != nil {
		return nil, err
	}

	next := current.Clone()
	next.SetBoard(board)
	next.SetTurnNumber(current.TurnNumber() + 1)

	if err = a.gameService.ValidateBoard(current.GameBoard, next); err != nil {
		return nil, err
	}

	nextCg := models.NewCurrentGameWithID(id, next)

	if _, ok := a.gameService.IsEnded(board); ok {
		if err = a.repository.Save(*nextCg); err != nil {
			return nil, err
		}
		return nextCg, nil
	}

	next = a.gameService.GetNextTurn(nextCg.GameBoard)
	if next == nil {
		return nil, errors.New("failed to calculate next turn")
	}
	nextCg.GameBoard = next

	if err = a.repository.Save(*nextCg); err != nil {
		return nil, err
	}

	return nextCg, nil
}

func (a appService) GetGame(id uuid.UUID) (*models.CurrentGame, error) {
	cg, err := a.repository.Get(id)
	if err != nil {
		return nil, err
	}

	return cg, nil
}
