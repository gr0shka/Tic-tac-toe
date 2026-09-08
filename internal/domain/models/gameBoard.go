package models

import "sync"

const (
	EmptyCell = -1

	CountPlayers = 2
	FirstPlayer  = 0
	SecondPlayer = 1
	BoardSize    = 3
)

type GameBoard struct {
	board        [BoardSize][BoardSize]int
	numberOfTurn int
	players      [CountPlayers]Player
	mu           sync.RWMutex
}

func NewGameBoard() *GameBoard {
	board := [BoardSize][BoardSize]int{}

	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			board[i][j] = EmptyCell
		}
	}

	return &GameBoard{
		board:        board,
		numberOfTurn: 0,
		mu:           sync.RWMutex{},
	}
}

func (gb *GameBoard) Clone() *GameBoard {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	clone := &GameBoard{
		board:        gb.board,
		numberOfTurn: gb.numberOfTurn,
		players:      gb.players,
		mu:           sync.RWMutex{},
	}
	return clone
}

func (gb *GameBoard) AddPlayers(player1, player2 Player) {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	gb.players[FirstPlayer] = player1
	gb.players[SecondPlayer] = player2
}

func (gb *GameBoard) SetTurnNumber(turnNumber int) {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	gb.numberOfTurn = turnNumber
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

func (gb *GameBoard) GetBoard() [BoardSize][BoardSize]int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return gb.board
}

func (gb *GameBoard) GetPlayers() [CountPlayers]Player {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	return gb.players
}

func (gb *GameBoard) NextPlayer() (Player, bool) {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	turnNumber := gb.numberOfTurn % CountPlayers

	if turnNumber >= len(gb.players) {
		return Player{}, false
	}

	return gb.players[gb.numberOfTurn%CountPlayers], true
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

func (gb *GameBoard) SetBoard(board [BoardSize][BoardSize]int) {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	gb.board = board
}
