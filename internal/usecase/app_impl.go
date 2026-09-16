package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/service"
)

type appService struct {
	gameService service.GameService
	repository  Repository
}

func NewAppService(gameService service.GameService, rep Repository) *appService {
	return &appService{
		gameService: gameService,
		repository:  rep,
	}
}

func (a appService) GameIsEnded(ctx context.Context, id uuid.UUID) (int, bool) {
	cg, err := a.repository.Get(id)
	if err != nil {
		return 0, false
	}

	return a.gameService.IsEnded(cg.Board())
}

func (a appService) CreateGame(ctx context.Context) (*game.CurrentGame, error) {
	gb := game.NewGameBoard()
	player := game.NewPlayer(game.FirstPlayer, true)
	computer := game.NewPlayer(game.SecondPlayer, false)
	gb.AddPlayers(*player, *computer)

	cg := game.NewCurrentGame(gb)

	if err := a.repository.Save(cg); err != nil {
		return nil, err
	}

	return cg, nil
}

func (a appService) ProcessPlayerMove(
	ctx context.Context,
	id uuid.UUID,
	board [game.BoardSize][game.BoardSize]int,
) (*game.CurrentGame, error) {

	current, err := a.repository.Get(id)
	if err != nil {
		return nil, err
	}

	if current.IsEnded() {
		return current, nil
	}

	next := current.Clone()
	next.SetBoard(board)
	next.SetTurnNumber(current.TurnNumber() + 1)

	if err = a.gameService.ValidateBoard(current.GameBoard, next); err != nil {
		return nil, err
	}

	nextCg := game.NewCurrentGameWithID(id, next)

	winner, ended := a.gameService.IsEnded(nextCg.Board())
	nextCg.SetWinner(winner)
	nextCg.SetIsEnded(ended)

	if ended {
		if err = a.repository.Save(nextCg); err != nil {
			return nil, err
		}
		return nextCg, nil
	}

	next = a.gameService.GetNextTurn(ctx, nextCg.GameBoard)
	if next == nil {
		return nil, game.ErrFailedCalculateNextTurn
	}
	nextCg.GameBoard = next

	winner, ended = a.gameService.IsEnded(nextCg.Board())
	nextCg.SetWinner(winner)
	nextCg.SetIsEnded(ended)

	if err = a.repository.Save(nextCg); err != nil {
		return nil, err
	}

	return nextCg, nil
}

func (a appService) GetGame(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	cg, err := a.repository.Get(id)
	if err != nil {
		return nil, err
	}

	return cg, nil
}
