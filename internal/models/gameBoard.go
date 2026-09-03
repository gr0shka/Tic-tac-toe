package models

import "sync"

type GameBoard struct {
	size     int
	board    [][]int
	nextTurn *Player
	mu       sync.RWMutex
}

func NewGameBoard(size int) *GameBoard {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)
	}

	return &GameBoard{
		size:  size,
		board: board,
		mu:    sync.RWMutex{},
	}
}

func (gb *GameBoard) Size() int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return gb.size
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
