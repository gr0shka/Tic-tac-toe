package models

import "sync"

const (
	EmptyCell = -1

	CountPlayers = 2
	boardSize    = 3
)

type GameBoard struct {
	board        [boardSize][boardSize]int
	numberOfTurn int
	players      []*Player
	mu           sync.RWMutex
}

func NewGameBoard() *GameBoard {
	board := [boardSize][boardSize]int{}

	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			board[i][j] = EmptyCell
		}
	}

	return &GameBoard{
		board:        board,
		numberOfTurn: 0,
		players:      make([]*Player, 0),
		mu:           sync.RWMutex{},
	}
}

func (gb *GameBoard) AddPlayers(player ...*Player) {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	for _, p := range player {
		gb.players = append(gb.players, p)
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

func (gb *GameBoard) GetBoard() [boardSize][boardSize]int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return gb.board
}

func (gb *GameBoard) HowIsNextTurn() *Player {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return gb.players[gb.numberOfTurn%CountPlayers]
}

func (gb *GameBoard) NextTurn() {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	gb.numberOfTurn++
}

func (gb *GameBoard) GetNumberOfTurns() int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return gb.numberOfTurn
}

func (gb *GameBoard) SetBoard(board [boardSize][boardSize]int) {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	gb.board = board
}
