package mapper

import (
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
)

func ToGameBoardResponse(cg *game.CurrentGame, winner int, ended bool) dto.GameBoardResponse {
	return dto.GameBoardResponse{
		ID:          cg.ID(),
		GameIsEnded: ended,
		Winner:      winner,
		Board:       cg.Board(),
	}
}
