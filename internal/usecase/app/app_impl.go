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

func (a appService) CreateGameWithBot(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	gb := game.NewGameBoard()
	player := game.NewPlayer(id, game.FirstPlayer, true)
	computer := game.NewPlayer(uuid.New(), game.SecondPlayer, false)
	gb.AddPlayers(player, computer)

	cg := game.NewCurrentGame(gb)

	cg.SetActivePlayer(player)
	cg.SetStatus(game.StatusPlayerTurn)

	if err := a.repository.Save(ctx, cg); err != nil {
		return nil, err
	}

	return cg, nil
}

func (a appService) CreateGameWithPlayer(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	gb := game.NewGameBoard()
	player := game.NewPlayer(id, game.FirstPlayer, true)
	gb.AddPlayers(player, nil)

	cg := game.NewCurrentGame(gb)

	cg.SetActivePlayer(nil)
	cg.SetStatus(game.StatusWaitingForPlayers)

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

	nextCg = a.checkTurn(nextCg)

	if nextCg.ActivePlayer() != nil && !nextCg.ActivePlayer().IsRealPlayer() {
		if nextCg, err = a.BotTurn(ctx, nextCg); err != nil {
			return nil, err
		}

		nextCg = a.checkTurn(nextCg)
	}

	if err = a.repository.Update(ctx, nextCg); err != nil {
		return nil, err
	}

	return nextCg, nil
}

func (a appService) checkTurn(cg *game.CurrentGame) *game.CurrentGame {
	if cg.IsEnded() {
		if cg.Status() == game.StatusDraw {
			cg.SetActivePlayer(nil)
		}

		return cg
	}

	if player, ok := cg.NextPlayer(); ok {
		cg.SetActivePlayer(player)
	}

	return cg
}

func (a appService) BotTurn(ctx context.Context, cg *game.CurrentGame) (*game.CurrentGame, error) {
	next := a.gameService.GetNextTurn(ctx, cg.GameBoard)
	if next == nil {
		return nil, game.ErrFailedCalculateNextTurn
	}
	cg.GameBoard = next

	winner, ended := a.gameService.IsEnded(cg.Board())
	cg.SetWinner(winner)
	cg.SetStatus(ended)

	return cg, nil
}

func (a appService) GetGame(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	cg, err := a.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return cg, nil
}
