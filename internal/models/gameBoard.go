package models

import "sync"

const (
	EmptyCage = -1

	countPlayers = 2
	boardSize    = 3
)

type GameBoard struct {
	board        [][]int
	numberOfTurn int
	mu           sync.RWMutex
}

func NewGameBoard() *GameBoard {
	board := make([][]int, boardSize)
	for i := 0; i < boardSize; i++ {
		board[i] = make([]int, boardSize)

		for j := 0; j < boardSize; j++ {
			board[i][j] = EmptyCage
		}
	}

	return &GameBoard{
		board:        board,
		numberOfTurn: 0,
		mu:           sync.RWMutex{},
	}
}

func (gb *GameBoard) Size() int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return boardSize
}

func (gb *GameBoard) Set(x, y int, v int) {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	gb.board[x][y] = v
}

func (gb *GameBoard) Get(x, y int) int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return gb.board[x][y]
}

func (gb *GameBoard) GetBoard() [][]int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return gb.board
}

func (gb *GameBoard) HowIsNextTurn() int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return gb.numberOfTurn % countPlayers
}

func (gb *GameBoard) NextTurn() {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	gb.numberOfTurn++
}
