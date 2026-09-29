package mapper

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/utils"
)

func ToGameBoardResponse(cg *game.CurrentGame) dto.GameBoardResponse {
	var p2Symbol *int
	if len(cg.Players()) > 1 && cg.Players()[game.SecondPlayer] != nil {
		p2Symbol = utils.Ptr(cg.Players()[game.SecondPlayer].Symbol())
	}

	var activePlayer *uuid.UUID
	if cg.ActivePlayer() != nil {
		activePlayer = utils.Ptr(cg.ActivePlayer().ID())
	}

	return dto.GameBoardResponse{
		ID:            cg.ID(),
		Status:        string(cg.Status()),
		ActivePlayer:  activePlayer,
		Player1Symbol: cg.Players()[game.FirstPlayer].Symbol(),
		Player2Symbol: p2Symbol,
		Winner:        cg.Winner(),
		Board:         cg.Board(),
	}
}
