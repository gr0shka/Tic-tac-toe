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

func (a appService) JoinGame(ctx context.Context, gameID, playerID uuid.UUID) (*game.CurrentGame, error) {
	cg, err := a.repository.Get(ctx, gameID)
	if err != nil {
		return nil, err
	}

	player := game.NewPlayer(playerID, game.SecondPlayer, true)

	if err = cg.AddPlayer(player); err != nil {
		return nil, err
	}

	turnPlayer, err := cg.NextPlayer()
	if err != nil {
		return nil, err
	}

	cg.SetActivePlayer(turnPlayer)
	cg.SetStatus(game.StatusPlayerTurn)

	if err = a.repository.Save(ctx, cg); err != nil {
		return nil, err
	}

	return cg, nil
}

func (a appService) AllGames(ctx context.Context) ([]*game.CurrentGame, error) {
	res, err := a.repository.AllGames(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
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

func (a appService) CreateGame(ctx context.Context, id uuid.UUID, mode game.GameMode) (*game.CurrentGame, error) {
	switch mode {
	case game.GameModePlayerVSPlayer:
		return a.CreateGameWithPlayer(ctx, id)
	case game.GameModePlayerVSBot:
		return a.CreateGameWithBot(ctx, id)

	default:
		return nil, game.ErrNonExistentMode
	}
}

func (a appService) CreateGameWithBot(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	gb := game.NewGameBoard()

	player := game.NewPlayer(id, game.FirstPlayer, true)
	if err := gb.AddPlayer(player); err != nil {
		return nil, err
	}

	computer := game.NewPlayer(uuid.New(), game.SecondPlayer, false)
	if err := gb.AddPlayer(computer); err != nil {
		return nil, err
	}

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
	if err := gb.AddPlayer(player); err != nil {
		return nil, err
	}

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
	playerID, gameID uuid.UUID,
	board [game.BoardSize][game.BoardSize]int,
) (*game.CurrentGame, error) {

	current, err := a.repository.Get(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if current.IsEnded() {
		return current, nil
	}

	if current.ActivePlayer() == nil || current.ActivePlayer().ID() != playerID {
		return nil, game.ErrWrongPlayerMove
	}

	next := current.Clone()
	next.SetBoard(board)
	next.SetTurnNumber(current.TurnNumber() + 1)

	if err = a.gameService.ValidateBoard(current.GameBoard, next); err != nil {
		return nil, err
	}

	nextCg := game.NewCurrentGameWithID(gameID, next)

	winner, ended := a.gameService.IsEnded(nextCg.Board())
	nextCg.SetWinner(winner)
	nextCg.SetStatus(ended)

	nextCg, err = a.checkTurn(nextCg)
	if err != nil {
		return nil, err
	}

	if nextCg.ActivePlayer() != nil && !nextCg.ActivePlayer().IsRealPlayer() {
		if nextCg, err = a.botTurn(ctx, nextCg); err != nil {
			return nil, err
		}

		nextCg, err = a.checkTurn(nextCg)
		if err != nil {
			return nil, err
		}
	}

	if err = a.repository.Update(ctx, nextCg); err != nil {
		return nil, err
	}

	return nextCg, nil
}

func (a appService) checkTurn(cg *game.CurrentGame) (*game.CurrentGame, error) {
	if cg.IsEnded() {
		if cg.Status() == game.StatusDraw {
			cg.SetActivePlayer(nil)
		}

		return cg, nil
	}

	player, err := cg.NextPlayer()
	if err != nil {
		return nil, err
	}

	cg.SetActivePlayer(player)

	return cg, nil
}

func (a appService) botTurn(ctx context.Context, cg *game.CurrentGame) (*game.CurrentGame, error) {
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
