package mapper

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
)

func ToGameBoardResponse(cg *game.CurrentGame) dto.GameBoardResponse {
	var p2ID *uuid.UUID
	var p2Symbol *int
	if len(cg.Players()) > 1 && cg.Players()[game.SecondPlayer] != nil {
		p2ID = new(cg.Players()[game.SecondPlayer].ID())
		p2Symbol = new(cg.Players()[game.SecondPlayer].Symbol())
	}

	var activePlayer *uuid.UUID
	if cg.ActivePlayer() != nil {
		activePlayer = new(cg.ActivePlayer().ID())
	}

	var winnerID *uuid.UUID
	if cg.Status() == game.StatusPlayerWins {
		winnerID = new(cg.Players()[cg.Winner()].ID())
	}

	return dto.GameBoardResponse{
		ID:           cg.ID(),
		Status:       string(cg.Status()),
		ActivePlayer: activePlayer,

		Player1ID:     cg.Players()[game.FirstPlayer].ID(),
		Player1Symbol: cg.Players()[game.FirstPlayer].Symbol(),

		Player2Symbol: p2Symbol,
		Player2ID:     p2ID,

		Winner:   cg.Winner(),
		WinnerID: winnerID,
		Board:    cg.Board(),
	}
}
