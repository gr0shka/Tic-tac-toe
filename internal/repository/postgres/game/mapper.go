package game

import (
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

func DomainToDTO(cg game.CurrentGame) CurrentGameDTO {
	cgPlayers := cg.Players()

	return CurrentGameDTO{
		ID:           cg.ID(),
		Board:        FlattenBoard(cg.Board()),
		NumberOfTurn: cg.TurnNumber(),
		Player1ID:    cgPlayers[0].ID(),
		Player1Real:  cgPlayers[0].IsRealPlayer(),
		Player2ID:    cgPlayers[1].ID(),
		Player2Real:  cgPlayers[1].IsRealPlayer(),
		Status:       string(cg.Status()),
		Winner:       cg.Winner(),
	}
}

func DTOToDomain(cgd CurrentGameDTO) *game.CurrentGame {
	gameBoard := game.NewGameBoard()
	gameBoard.SetBoard(UnFlattenBoard(cgd.Board))
	gameBoard.SetTurnNumber(cgd.NumberOfTurn)

	player1 := game.NewPlayer(
		cgd.Player1ID,
		0,
		cgd.Player1Real,
	)
	player2 := game.NewPlayer(
		cgd.Player2ID,
		1,
		cgd.Player2Real,
	)
	gameBoard.AddPlayers(*player1, *player2)

	cg := game.NewCurrentGameWithID(cgd.ID, gameBoard)

	cg.SetStatus(game.GameStatus(cgd.Status))
	cg.SetWinner(cgd.Winner)

	return cg
}

func FlattenBoard(b [game.BoardSize][game.BoardSize]int) []int {
	res := make([]int, 0, game.BoardSize*game.BoardSize)

	for i := 0; i < game.BoardSize; i++ {
		for j := 0; j < game.BoardSize; j++ {
			res = append(res, b[i][j])
		}
	}

	return res
}

func UnFlattenBoard(b []int) [game.BoardSize][game.BoardSize]int {
	res := [game.BoardSize][game.BoardSize]int{}

	index := 0
	for i := 0; i < game.BoardSize; i++ {
		for j := 0; j < game.BoardSize; j++ {
			res[i][j] = b[index]
			index++
		}
	}

	return res
}
