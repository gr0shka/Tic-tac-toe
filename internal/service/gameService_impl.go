package service

import (
	"Tic-Tac-Toy/internal/models"
	"errors"
	"math"
)

const (
	Draw = -1
)

type repository interface {
}

type gameService struct {
	rep *repository
}

func NewGameService(rep *repository) *gameService {
	return &gameService{
		rep: rep,
	}
}

func (g gameService) GetNextTurn(gb models.GameBoard) *models.GameBoard {
	bestTurn := struct {
		score int
		x, y  int
	}{
		score: math.MinInt,
		x:     0,
		y:     0,
	}

	for i := 0; i < gb.Size(); i++ {
		for j := 0; j < gb.Size(); j++ {

			if gb.Get(i, j) == models.EmptyCage {
				tempGb := gb
				turneNum := tempGb.HowIsNextTurn().GetTurnNumber()
				tempGb.Set(i, j, turneNum)

				if sc := g.recursiveScoring(tempGb); sc > bestTurn.score {
					bestTurn.score = sc
					bestTurn.x = i
					bestTurn.y = j
				}
			}
		}
	}

	gb.Set(bestTurn.x, bestTurn.y, gb.HowIsNextTurn().GetTurnNumber())
	gb.NextTurn()

	return &gb
}

func (g gameService) recursiveScoring(gb models.GameBoard) int {
	score := 0

	if gb.HowIsNextTurn().IsRealPlayer() {
		score = math.MaxInt
	} else {
		score = math.MinInt
	}

	p, end := g.IsEnded(gb.GetBoard())
	if end {
		if !gb.HowIsNextTurn().IsRealPlayer() {

			if gb.HowIsNextTurn().GetTurnNumber() == p {
				return 1
			}
			if p != Draw {
				return -1
			}

		} else if gb.HowIsNextTurn().GetTurnNumber() == p {
			return -1
		}

		return 0
	}

	board := gb.GetBoard()
	gb.NextTurn()
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {

			if board[i][j] == models.EmptyCage {
				tempGb := gb
				tempGb.Set(i, j, gb.HowIsNextTurn().GetTurnNumber())

				if gb.HowIsNextTurn().IsRealPlayer() {
					score = min(score, g.recursiveScoring(tempGb))
				} else {
					score = max(score, g.recursiveScoring(tempGb))
				}
			}
		}
	}

	return score
}

func (g gameService) ValidateBoard(oldB, newB models.GameBoard) error {
	if newB.GetNumberOfTurns() != oldB.GetNumberOfTurns()+1 {
		return errors.New("wrong number of turns")
	}

	newBoard := newB.GetBoard()
	oldBoard := oldB.GetBoard()

	countChangedCage := 0
	for i := 0; i < len(newBoard); i++ {
		for j := 0; j < len(newBoard[i]); j++ {

			if newBoard[i][j] != oldBoard[i][j] {
				countChangedCage++
			}
		}
	}

	if countChangedCage != 1 {
		return errors.New("wrong number of turns")
	}

	return nil
}

func (g gameService) IsEnded(board [][]int) (int, bool) {

	checkBoard := func(board [][]int, fns ...func([][]int) (int, bool)) (int, bool) {
		for _, fn := range fns {
			cageType, ok := fn(board)
			if ok {
				return cageType, true
			}
		}

		return Draw, false
	}

	return checkBoard(board, horizontalCheck, verticalCheck, diagonalsCheck, allCageOccupied)
}

func horizontalCheck(board [][]int) (int, bool) {
	for i := 0; i < len(board); i++ {
		cageType := board[i][0]
		flag := true

		for j := 1; j < len(board[i]); j++ {

			if cageType != board[i][j] {
				flag = false
				break
			}
		}

		if flag {
			return cageType, true
		}
	}

	return Draw, false
}

func verticalCheck(board [][]int) (int, bool) {
	for j := 0; j < len(board); j++ {
		cageType := board[0][j]
		flag := true

		for i := 1; i < len(board[j]); i++ {

			if cageType != board[i][j] {
				flag = false
				break
			}
		}

		if flag {
			return cageType, true
		}
	}

	return Draw, false
}

func diagonalsCheck(board [][]int) (int, bool) {
	cageType := board[0][0]
	flag := true

	for i := 1; i < len(board); i++ {
		if cageType != board[i][i] {
			flag = false
			break
		}
	}
	if flag {
		return cageType, true
	}

	cageType = board[len(board)-1][0]
	flag = true

	for i := 1; i < len(board); i++ {
		if cageType != board[(len(board)-1)-i][i] {
			flag = false
			break
		}
	}
	if flag {
		return cageType, true
	}

	return Draw, false
}

func allCageOccupied(board [][]int) (int, bool) {

	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {

			if board[i][j] == models.EmptyCage {

				return Draw, false
			}
		}
	}

	return Draw, true
}
