package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/service"
)

type appService struct {
	gameService service.GameService
	repository  GameRepository
}

func NewAppService(gameService service.GameService, rep GameRepository) *appService {
	return &appService{
		gameService: gameService,
		repository:  rep,
	}
}

func (a appService) GameIsEnded(ctx context.Context, id uuid.UUID) (int, game.GameStatus) {
	cg, err := a.repository.Get(ctx, id)
	if err != nil {
		return 0, game.StatusDraw
	}

	return a.gameService.IsEnded(cg.Board())
}

func (a appService) CreateGame(ctx context.Context) (*game.CurrentGame, error) {
	gb := game.NewGameBoard()
	player := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
	computer := game.NewPlayer(uuid.New(), game.SecondPlayer, false)
	gb.AddPlayers(*player, *computer)

	cg := game.NewCurrentGame(gb)

	if err := a.repository.Save(ctx, cg); err != nil {
		return nil, err
	}

	return cg, nil
}

func (a appService) ProcessPlayerMove(
	ctx context.Context,
	id uuid.UUID,
	board [game.BoardSize][game.BoardSize]int,
) (*game.CurrentGame, error) {

	current, err := a.repository.Get(ctx, id)
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
	nextCg.SetStatus(ended)

	if ended == game.StatusPlayerWins || ended == game.StatusDraw {
		if err = a.repository.Update(ctx, nextCg); err != nil {
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
	nextCg.SetStatus(ended)

	if err = a.repository.Update(ctx, nextCg); err != nil {
		return nil, err
	}

	return nextCg, nil
}

func (a appService) GetGame(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	cg, err := a.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return cg, nil
}
