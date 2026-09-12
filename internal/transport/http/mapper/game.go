package mapper

func ToGameBoardResponse(cg *game.CurrentGame, winner int, ended bool) dto.GameBoardResponse {
	return dto.GameBoardResponse{
		ID:          cg.ID(),
		GameIsEnded: ended,
		Winner:      winner,
		Board:       cg.Board(),
	}
}
