package mapstore

import (
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
)

func DomainToDTO(cg models.CurrentGame) CurrentGameDTO {
	cgPlayers := cg.GetPlayers()
	players := [models.CountPlayers]PlayerDTO{
		{cgPlayers[models.FirstPlayer].GetTurnNumber(), cgPlayers[models.FirstPlayer].IsRealPlayer()},
		{cgPlayers[models.SecondPlayer].GetTurnNumber(), cgPlayers[models.SecondPlayer].IsRealPlayer()},
	}

	return CurrentGameDTO{
		ID:           cg.ID(),
		Board:        cg.GetBoard(),
		NumberOfTurn: cg.GetNumberOfTurns(),
		Players:      players,
	}
}

func DTOToDomain(cgd CurrentGameDTO) *models.CurrentGame {
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
	gameBoard.AddPlayers(*player1, *player2)

	return models.NewCurrentGameWithID(cgd.ID, gameBoard)
}
