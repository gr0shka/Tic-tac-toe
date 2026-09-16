package service

import (
	"context"
	"math"

	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

const (
	Draw     = -1
	MaxDepth = 10
)

type gameService struct {
}

func NewGameService() *gameService {
	return &gameService{}
}

func (g gameService) GetNextTurn(ctx context.Context, gb *game.GameBoard) *game.GameBoard {
	bestTurn := struct {
		score int
		x, y  int
	}{
		score: math.MinInt,
		x:     -1,
		y:     -1,
	}

	for i := 0; i < game.BoardSize; i++ {
		for j := 0; j < game.BoardSize; j++ {
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			if gb.Get(i, j) != game.EmptyCell {
				continue
			}

			tempGb := gb.Clone()
			p, ok := tempGb.NextPlayer()
			if !ok {
				continue
			}
			tempGb.Set(i, j, p.Symbol())
			tempGb.NextTurn()

			depth := 0

			if sc := g.recursiveScoring(ctx, *tempGb, depth); sc > bestTurn.score {
				bestTurn.score = sc
				bestTurn.x = i
				bestTurn.y = j
			}
		}
	}

	np, ok := gb.NextPlayer()
	if !ok {
		return nil
	}

	if bestTurn.x == -1 || bestTurn.y == -1 {
		return nil
	}

	gb.Set(bestTurn.x, bestTurn.y, np.Symbol())
	gb.NextTurn()

	return gb
}

func (g gameService) recursiveScoring(ctx context.Context, gb game.GameBoard, depth int) int {
	current, ok := gb.NextPlayer()
	if !ok {
		return 0
	}

	score := math.MinInt
	if current.IsRealPlayer() {
		score = math.MaxInt
	}

	p, end := g.IsEnded(gb.Board())
	if end {
		if p == Draw {
			return 0
		}
		if p == game.FirstPlayer {
			return -(MaxDepth - depth)
		}
		return MaxDepth - depth
	}

	board := gb.Board()
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {
			select {
			case <-ctx.Done():
				return 0
			default:
			}

			if board[i][j] != game.EmptyCell {
				continue
			}

			branch := gb.Clone()
			cp, ok := branch.NextPlayer()
			if !ok {
				continue
			}

			branch.Set(i, j, cp.Symbol())
			branch.NextTurn()

			depth++

			if current.IsRealPlayer() {
				score = min(score, g.recursiveScoring(ctx, *branch, depth))
			} else {
				score = max(score, g.recursiveScoring(ctx, *branch, depth))
			}
		}
	}

	return score
}

func (g gameService) ValidateBoard(oldB, newB *game.GameBoard) error {
	if newB.TurnNumber() != oldB.TurnNumber()+1 {
		return game.ErrWrongNumberOfTurns
	}

	newBoard := newB.Board()
	oldBoard := oldB.Board()

	player, ok := oldB.NextPlayer()
	if !ok {
		return game.ErrInvalidTurn
	}

	changedX, changedY := -1, -1
	for i := 0; i < game.BoardSize; i++ {
		for j := 0; j < game.BoardSize; j++ {
			if newBoard[i][j] != oldBoard[i][j] {

				if oldBoard[i][j] != game.EmptyCell {
					return game.ErrAlteredPreviousMoves
				}

				if changedX != -1 || changedY != -1 {
					return game.ErrMultipleMoves
				}

				changedX, changedY = i, j
			}
		}
	}

	if changedX == -1 || changedY == -1 {
		return game.ErrNoMoveMade
	}

	if newBoard[changedX][changedY] != player.Symbol() {
		return game.ErrWrongPlayerMove
	}

	return nil
}

func (g gameService) IsEnded(board [game.BoardSize][game.BoardSize]int) (int, bool) {

	checkBoard := func(board [game.BoardSize][game.BoardSize]int, fns ...func([game.BoardSize][game.BoardSize]int) (int, bool)) (int, bool) {
		for _, fn := range fns {
			cellType, ok := fn(board)
			if ok {
				return cellType, true
			}
		}

		return Draw, false
	}

	return checkBoard(board, horizontalCheck, verticalCheck, diagonalsCheck, allCellOccupied)
}

func horizontalCheck(board [game.BoardSize][game.BoardSize]int) (int, bool) {
	for i := 0; i < len(board); i++ {
		cellType := board[i][0]
		flag := true

		if cellType == game.EmptyCell {
			flag = false
			continue
		}

		for j := 1; j < len(board[i]); j++ {

			if cellType != board[i][j] {
				flag = false
				break
			}
		}

		if flag {
			return cellType, true
		}
	}

	return Draw, false
}

func verticalCheck(board [game.BoardSize][game.BoardSize]int) (int, bool) {
	for j := 0; j < len(board); j++ {
		cellType := board[0][j]
		flag := true

		if cellType == game.EmptyCell {
			flag = false
			continue
		}

		for i := 1; i < len(board[j]); i++ {

			if cellType != board[i][j] {
				flag = false
				break
			}
		}

		if flag {
			return cellType, true
		}
	}

	return Draw, false
}

func diagonalsCheck(board [game.BoardSize][game.BoardSize]int) (int, bool) {
	cellType := board[0][0]
	flag := true

	if cellType != game.EmptyCell {
		for i := 1; i < len(board); i++ {
			if cellType != board[i][i] {
				flag = false
				break
			}
		}

	} else {
		flag = false
	}

	if flag {
		return cellType, true
	}

	cellType = board[len(board)-1][0]
	flag = true
	if cellType != game.EmptyCell {
		for i := 1; i < len(board); i++ {
			if cellType != board[(len(board)-1)-i][i] {
				flag = false
				break
			}
		}

	} else {
		flag = false
	}

	if flag {
		return cellType, true
	}

	return Draw, false
}

func allCellOccupied(board [game.BoardSize][game.BoardSize]int) (int, bool) {

	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {

			if board[i][j] == game.EmptyCell {

				return Draw, false
			}
		}
	}

	return Draw, true
}
