package dto

type GameBoardResponse struct {
	GameIsEnded bool
	HowIsWinner int
	Board       [boardSize][boardSize]int `json:"board"`
}
