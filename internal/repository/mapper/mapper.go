package mapper

import (
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/DTO"
)

func DomainToDTO(cg models.CurrentGame) DTO.CurrentGameDTO {
	cgPlayers := cg.GetPlayers()
	players := [models.CountPlayers]DTO.PlayerDTO{
		{cgPlayers[models.FirstPlayer].GetTurnNumber(), cgPlayers[models.FirstPlayer].IsRealPlayer()},
		{cgPlayers[models.SecondPlayer].GetTurnNumber(), cgPlayers[models.SecondPlayer].IsRealPlayer()},
	}

	return DTO.CurrentGameDTO{
		ID:           cg.ID(),
		Board:        cg.GetBoard(),
		NumberOfTurn: cg.GetNumberOfTurns(),
		Players:      players,
	}
}

func DTOToDomain(cgd DTO.CurrentGameDTO) models.CurrentGame {
	gameBoard := models.NewGameBoard()
	gameBoard.SetBoard(cgd.Board)
	gameBoard.SetTurnNumber(cgd.NumberOfTurn)

	player1 := models.NewPlayer(
		cgd.Players[models.FirstPlayer].TurnNumber,
		cgd.Players[models.FirstPlayer].RealPlayer,
	)
	player2 := models.NewPlayer(
		cgd.Players[models.SecondPlayer].TurnNumber,
		cgd.Players[models.SecondPlayer].RealPlayer,
	)
	gameBoard.AddPlayers(player1, player2)

	return models.NewCurrentGameWithID(cgd.ID, gameBoard)
}
