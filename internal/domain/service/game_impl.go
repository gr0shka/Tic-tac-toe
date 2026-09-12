package service

import (
	"errors"
	"math"

	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
)

const (
	Draw = -1
)

type gameService struct {
}

func NewGameService() *gameService {
	return &gameService{}
}

func (g gameService) GetNextTurn(gb *models.GameBoard) *models.GameBoard {
	bestTurn := struct {
		score int
		x, y  int
	}{
		score: math.MinInt,
		x:     0,
		y:     0,
	}

	for i := 0; i < models.BoardSize; i++ {
		for j := 0; j < models.BoardSize; j++ {
			if gb.Get(i, j) != models.EmptyCell {
				continue
			}

			tempGb := gb.Clone()
			p, ok := tempGb.NextPlayer()
			if !ok {
				continue
			}
			tempGb.Set(i, j, p.TurnNumber())
			tempGb.NextTurn()

			if sc := g.recursiveScoring(*tempGb); sc > bestTurn.score {
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

	gb.Set(bestTurn.x, bestTurn.y, np.TurnNumber())
	gb.NextTurn()

	return gb
}

func (g gameService) recursiveScoring(gb models.GameBoard) int {
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
		if p == models.FirstPlayer {
			return -1
		}
		return 1
	}

	board := gb.Board()
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {
			if board[i][j] != models.EmptyCell {
				continue
			}

			branch := gb.Clone()
			cp, ok := branch.NextPlayer()
			if !ok {
				continue
			}

			branch.Set(i, j, cp.TurnNumber())
			branch.NextTurn()

			if current.IsRealPlayer() {
				score = min(score, g.recursiveScoring(*branch))
			} else {
				score = max(score, g.recursiveScoring(*branch))
			}
		}
	}

	return score
}

// Добавить больше проверок (правильный ли ход и тд)
func (g gameService) ValidateBoard(oldB, newB *models.GameBoard) error {
	if newB.TurnNumber() != oldB.TurnNumber()+1 {
		return errors.New("wrong number of turns")
	}

	newBoard := newB.Board()
	oldBoard := oldB.Board()

	countChangedCell := 0
	for i := 0; i < len(newBoard); i++ {
		for j := 0; j < len(newBoard[i]); j++ {

			if newBoard[i][j] != oldBoard[i][j] {
				countChangedCell++
			}
		}
	}

	if countChangedCell != 1 {
		return errors.New("wrong number of turns")
	}

	return nil
}

func (g gameService) IsEnded(board [models.BoardSize][models.BoardSize]int) (int, bool) {

	checkBoard := func(board [models.BoardSize][models.BoardSize]int, fns ...func([models.BoardSize][models.BoardSize]int) (int, bool)) (int, bool) {
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

func horizontalCheck(board [models.BoardSize][models.BoardSize]int) (int, bool) {
	for i := 0; i < len(board); i++ {
		cellType := board[i][0]
		flag := true

		if cellType == models.EmptyCell {
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

func verticalCheck(board [models.BoardSize][models.BoardSize]int) (int, bool) {
	for j := 0; j < len(board); j++ {
		cellType := board[0][j]
		flag := true

		if cellType == models.EmptyCell {
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

func diagonalsCheck(board [models.BoardSize][models.BoardSize]int) (int, bool) {
	cellType := board[0][0]
	flag := true

	if cellType != models.EmptyCell {
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
	if cellType != models.EmptyCell {
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

func allCellOccupied(board [models.BoardSize][models.BoardSize]int) (int, bool) {

	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {

			if board[i][j] == models.EmptyCell {

				return Draw, false
			}
		}
	}

	return Draw, true
}
