package game

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/utils"
)

func DomainToDTO(cg game.CurrentGame) CurrentGameDTO {
	cgPlayers := cg.Players()

	var p2ID *uuid.UUID
	var p2Real *bool
	var p2Symbol *int
	if len(cgPlayers) > 1 && cgPlayers[game.SecondPlayer] != nil {
		p2ID = utils.Ptr(cgPlayers[game.SecondPlayer].ID())
		p2Real = utils.Ptr(cgPlayers[game.SecondPlayer].IsRealPlayer())
		p2Symbol = utils.Ptr(cgPlayers[game.SecondPlayer].Symbol())
	}

	var activePlayerID *uuid.UUID
	if cg.ActivePlayer() != nil {
		activePlayerID = utils.Ptr(cg.ActivePlayer().ID())
	}

	return CurrentGameDTO{
		ID: cg.ID(),

		Player1ID:     cgPlayers[game.FirstPlayer].ID(),
		Player1Real:   cgPlayers[game.FirstPlayer].IsRealPlayer(),
		Player1Symbol: cgPlayers[game.FirstPlayer].Symbol(),

		Player2ID:     p2ID,
		Player2Real:   p2Real,
		Player2Symbol: p2Symbol,

		ActivePlayerID: activePlayerID,

		Board:        FlattenBoard(cg.Board()),
		NumberOfTurn: cg.TurnNumber(),
		Status:       string(cg.Status()),
		Winner:       cg.Winner(),
	}
}

func DTOToDomain(cgd CurrentGameDTO) *game.CurrentGame {
	gameBoard := game.NewGameBoard()
	gameBoard.SetBoard(UnFlattenBoard(cgd.Board))
	gameBoard.SetTurnNumber(cgd.NumberOfTurn)

	var activePlayer *game.Player

	player1 := game.NewPlayer(
		cgd.Player1ID,
		cgd.Player1Symbol,
		cgd.Player1Real,
	)
	gameBoard.AddPlayer(player1)

	if cgd.ActivePlayerID != nil && *cgd.ActivePlayerID == player1.ID() {
		activePlayer = player1
	}

	if cgd.Player2ID != nil {
		player2 := game.NewPlayer(
			*cgd.Player2ID,
			*cgd.Player2Symbol,
			*cgd.Player2Real,
		)
		gameBoard.AddPlayer(player2)

		if cgd.ActivePlayerID != nil && *cgd.ActivePlayerID == player2.ID() {
			activePlayer = player2
		}
	}

	cg := game.NewCurrentGameWithID(cgd.ID, gameBoard)

	cg.SetActivePlayer(activePlayer)

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
