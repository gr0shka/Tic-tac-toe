package mapstore

import (
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

func DomainToDTO(cg game.CurrentGame) CurrentGameDTO {
	cgPlayers := cg.Players()
	players := [game.CountPlayers]PlayerDTO{
		{cgPlayers[game.FirstPlayer].TurnNumber(), cgPlayers[game.FirstPlayer].IsRealPlayer()},
		{cgPlayers[game.SecondPlayer].TurnNumber(), cgPlayers[game.SecondPlayer].IsRealPlayer()},
	}

	return CurrentGameDTO{
		ID:           cg.ID(),
		Board:        cg.Board(),
		NumberOfTurn: cg.TurnNumber(),
		Players:      players,
	}
}

func DTOToDomain(cgd CurrentGameDTO) *game.CurrentGame {
	gameBoard := game.NewGameBoard()
	gameBoard.SetBoard(cgd.Board)
	gameBoard.SetTurnNumber(cgd.NumberOfTurn)

	player1 := game.NewPlayer(
		cgd.Players[game.FirstPlayer].TurnNumber,
		cgd.Players[game.FirstPlayer].RealPlayer,
	)
	player2 := game.NewPlayer(
		cgd.Players[game.SecondPlayer].TurnNumber,
		cgd.Players[game.SecondPlayer].RealPlayer,
	)
	gameBoard.AddPlayers(*player1, *player2)

	return game.NewCurrentGameWithID(cgd.ID, gameBoard)
}
