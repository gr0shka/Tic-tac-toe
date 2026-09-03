package service

import (
	"Tic-Tac-Toy/internal/models"
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

func (g gameService) GetNextTurn(gb models.GameBoard) models.GameBoard {
	//TODO implement me
	panic("implement me")
}

func (g gameService) ValidateBoard(oldBoard, newBoard models.GameBoard) error {
	//TODO implement me
	panic("implement me")
}

func (g gameService) IsEnded(gb models.GameBoard) (int, bool) {
	board := gb.GetBoard()

	checkBoard := func(board [][]int, fns ...func([][]int) (int, bool)) (int, bool) {
		for _, fn := range fns {
			cageType, ok := fn(board)
			if ok {
				return cageType, true
			}
		}

		return models.EmptyCage, false
	}

	return checkBoard(board)
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

	return models.EmptyCage, false
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

	return models.EmptyCage, false
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

	return models.EmptyCage, false
}
