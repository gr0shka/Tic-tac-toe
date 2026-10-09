package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/service"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase"
)

type appService struct {
	gameService service.GameService
	repository  GameRepository
	tx          usecase.Transactor
}

func NewAppService(gameService service.GameService, rep GameRepository, tx usecase.Transactor) *appService {
	return &appService{
		gameService: gameService,
		repository:  rep,
		tx:          tx,
	}
}

func (a appService) JoinGame(ctx context.Context, gameID, playerID uuid.UUID) (*game.CurrentGame, error) {
	var resultCG *game.CurrentGame

	err := a.tx.Do(ctx, func(txCtx context.Context) error {
		cg, err := a.repository.Get(txCtx, gameID)
		if err != nil {
			return err
		}

		if len(cg.Players()) > 0 && cg.Players()[game.FirstPlayer] != nil {
			if cg.Players()[game.FirstPlayer].ID() == playerID {
				return game.ErrPlayerAlreadyExists
			}
		}

		player := game.NewPlayer(playerID, game.SecondPlayer, true)

		if err = cg.AddPlayer(player); err != nil {
			return err
		}

		turnPlayer, err := cg.NextPlayer()
		if err != nil {
			return err
		}

		cg.SetActivePlayer(turnPlayer)
		cg.SetStatus(game.StatusPlayerTurn)

		if err = a.repository.Update(txCtx, cg); err != nil {
			return err
		}

		resultCG = cg
		return nil
	})

	return resultCG, err
}

func (a appService) AllGames(ctx context.Context) ([]*game.CurrentGame, error) {
	res, err := a.repository.AllGames(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
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

	var resultCG *game.CurrentGame

	err := a.tx.Do(ctx, func(txCtx context.Context) error {

		current, err := a.repository.Get(txCtx, gameID)
		if err != nil {
			return err
		}

		if current.IsEnded() {
			resultCG = current
			return nil
		}

		if current.ActivePlayer() == nil || current.ActivePlayer().ID() != playerID {
			return game.ErrWrongPlayerMove
		}

		next := current.Clone()
		next.SetBoard(board)
		next.SetTurnNumber(current.TurnNumber() + 1)

		if err = a.gameService.ValidateBoard(current.GameBoard, next); err != nil {
			return err
		}

		nextCg := game.NewCurrentGameWithID(gameID, next)

		winner, ended := a.gameService.IsEnded(nextCg.Board())
		nextCg.SetWinner(winner)
		nextCg.SetStatus(ended)

		nextCg, err = a.checkTurn(nextCg)
		if err != nil {
			return err
		}

		if nextCg.ActivePlayer() != nil && !nextCg.ActivePlayer().IsRealPlayer() {
			if nextCg, err = a.botTurn(txCtx, nextCg); err != nil {
				return err
			}

			nextCg, err = a.checkTurn(nextCg)
			if err != nil {
				return err
			}
		}

		if err = a.repository.Update(txCtx, nextCg); err != nil {
			return err
		}

		resultCG = nextCg
		return nil
	})

	return resultCG, err
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
