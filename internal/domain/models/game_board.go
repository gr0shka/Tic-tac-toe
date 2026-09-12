package models

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
	}
}

func (gb *GameBoard) Clone() *GameBoard {
	clone := &GameBoard{
		board:        gb.board,
		numberOfTurn: gb.numberOfTurn,
		players:      gb.players,
	}
	return clone
}

func (gb *GameBoard) AddPlayers(player1, player2 Player) {
	gb.players[FirstPlayer] = player1
	gb.players[SecondPlayer] = player2
}

func (gb *GameBoard) SetTurnNumber(turnNumber int) {
	gb.numberOfTurn = turnNumber
}

func (gb *GameBoard) Set(x, y int, v int) {
	gb.board[x][y] = v
}

func (gb *GameBoard) Get(x, y int) int {
	return gb.board[x][y]
}

func (gb *GameBoard) Board() [BoardSize][BoardSize]int {
	return gb.board
}

func (gb *GameBoard) Players() [CountPlayers]Player {
	return gb.players
}

func (gb *GameBoard) NextPlayer() (Player, bool) {
	turnNumber := gb.numberOfTurn % CountPlayers

	if turnNumber >= len(gb.players) {
		return Player{}, false
	}

	return gb.players[gb.numberOfTurn%CountPlayers], true
}

func (gb *GameBoard) NextTurn() {
	gb.numberOfTurn++
}

func (gb *GameBoard) TurnNumber() int {
	return gb.numberOfTurn
}

func (gb *GameBoard) SetBoard(board [BoardSize][BoardSize]int) {
	gb.board = board
}
