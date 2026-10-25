package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/service"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase"
)

type gameService struct {
	gameRules  service.GameRules
	repository GameRepository
	tx         usecase.Transactor
}

func (a *gameService) GetLeaderBoard(ctx context.Context, count int) (*game.LeaderBoard, error) {
	if count < 1 {
		return nil, fmt.Errorf("%w: count users in leader board must be >= 1", game.ErrInvalidArgument)
	}

	lb, err := a.repository.GetLeaderBoard(ctx, count)
	if err != nil {
		return nil, fmt.Errorf("get leader board: %w", err)
	}

	return lb, nil
}

func (a *gameService) AllEndedGamesByUserID(ctx context.Context, userID uuid.UUID) ([]*game.CurrentGame, error) {
	endedGames, err := a.repository.AllEndedGames(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository.AllEndedGames: %w", err)
	}

	return endedGames, nil
}

func NewGameService(gameRules service.GameRules, rep GameRepository, tx usecase.Transactor) *gameService {
	return &gameService{
		gameRules:  gameRules,
		repository: rep,
		tx:         tx,
	}
}

func (a *gameService) JoinGame(ctx context.Context, gameID, playerID uuid.UUID) (*game.CurrentGame, error) {
	var resultCG *game.CurrentGame

	err := a.tx.Do(ctx, func(txCtx context.Context) error {
		cg, err := a.repository.Get(txCtx, gameID)
		if err != nil {
			return err
		}

		if len(cg.Players()) > 0 && cg.Players()[game.FirstPlayer] != nil {
			if cg.Players()[game.FirstPlayer].ID() == playerID {
				return fmt.Errorf("%w: player already exists in this game", game.ErrPlayerAlreadyInGame)
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

func (a *gameService) AllGames(ctx context.Context) ([]*game.CurrentGame, error) {
	res, err := a.repository.AllGames(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (a *gameService) GameIsEnded(ctx context.Context, id uuid.UUID) (int, game.GameStatus, error) {
	cg, err := a.repository.Get(ctx, id)
	if err != nil {
		return 0, game.StatusDraw, fmt.Errorf("repository.Get: %w", err)
	}

	player, status := a.gameRules.IsEnded(cg.Board())
	return player, status, nil
}

func (a *gameService) CreateGame(ctx context.Context, id uuid.UUID, mode game.GameMode) (*game.CurrentGame, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: gameID is nil", game.ErrInvalidArgument)
	}

	switch mode {
	case game.GameModePlayerVSPlayer:
		return a.CreateGameWithPlayer(ctx, id)
	case game.GameModePlayerVSBot:
		return a.CreateGameWithBot(ctx, id)

	default:
		return nil, fmt.Errorf("%w: invalid mode", game.ErrInvalidArgument)
	}
}

func (a *gameService) CreateGameWithBot(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
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

func (a *gameService) CreateGameWithPlayer(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
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

func (a *gameService) ProcessPlayerMove(
	ctx context.Context,
	playerID, gameID uuid.UUID,
	board [game.BoardSize][game.BoardSize]int,
) (*game.CurrentGame, error) {

	if playerID == uuid.Nil {
		return nil, fmt.Errorf("%w: playerID is nil", game.ErrInvalidArgument)
	}

	if gameID == uuid.Nil {
		return nil, fmt.Errorf("%w: gameID is nil", game.ErrInvalidArgument)
	}

	var resultCG *game.CurrentGame

	err := a.tx.Do(ctx, func(txCtx context.Context) error {

		current, err := a.repository.Get(txCtx, gameID)
		if err != nil {
			return err
		}

		if current.IsEnded() {
			resultCG = current
			return fmt.Errorf("%w: game is ended", game.ErrInvalidMove)
		}

		if current.ActivePlayer() == nil || current.ActivePlayer().ID() != playerID {
			return fmt.Errorf("%w: wrong player move got ID = %v, need ID = %v",
				game.ErrInvalidMove, playerID, current.ActivePlayer().ID())
		}

		next := current.Clone()
		next.SetBoard(board)
		next.SetTurnNumber(current.TurnNumber() + 1)

		if err = a.gameRules.ValidateBoard(current.GameBoard, next); err != nil {
			return err
		}

		nextCg := game.NewCurrentGameWithID(gameID, next)

		winner, ended := a.gameRules.IsEnded(nextCg.Board())
		nextCg.SetWinner(winner)
		nextCg.SetStatus(ended)
		nextCg.SetCreatedAt(current.CreatedAt())

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

func (a *gameService) checkTurn(cg *game.CurrentGame) (*game.CurrentGame, error) {
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

func (a *gameService) botTurn(ctx context.Context, cg *game.CurrentGame) (*game.CurrentGame, error) {
	next := a.gameRules.GetNextTurn(ctx, cg.GameBoard)
	if next == nil {
		return nil, fmt.Errorf("bot turn: failed calculating next turn")
	}
	cg.GameBoard = next

	winner, ended := a.gameRules.IsEnded(cg.Board())
	cg.SetWinner(winner)
	cg.SetStatus(ended)

	return cg, nil
}

func (a *gameService) GetGame(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	cg, err := a.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return cg, nil
}
