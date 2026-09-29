package mapper

import (
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/utils"
)

func ToGameBoardResponse(cg *game.CurrentGame) dto.GameBoardResponse {
	return dto.GameBoardResponse{
		ID:            cg.ID(),
		Status:        string(cg.Status()),
		ActivePlayer:  utils.Ptr(cg.ActivePlayer().ID()),
		Player1Symbol: cg.Players()[game.FirstPlayer].Symbol(),
		Player2Symbol: cg.Players()[game.SecondPlayer].Symbol(),
		Winner:        cg.Winner(),
		Board:         cg.Board(),
	}
}
