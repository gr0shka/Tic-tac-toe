package dto

const (
	boardSize = 3
)

type GameBoardRequest struct {
	Board [boardSize][boardSize]int `json:"board"`
}
