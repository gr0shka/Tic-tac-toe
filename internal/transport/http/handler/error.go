package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
)

type errorResponse struct {
	Error string `json:"error"`
}

func responseError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, game.ErrNotFound),
		errors.Is(err, user.ErrUserNotFound):

		status = http.StatusNotFound

	case errors.Is(err, game.ErrGameFull),
		errors.Is(err, game.ErrPlayerAlreadyInGame),
		errors.Is(err, user.ErrUserAlreadyExists):

		status = http.StatusConflict

	case errors.Is(err, game.ErrInvalidArgument),
		errors.Is(err, game.ErrInvalidMove),
		errors.Is(err, user.ErrInvalidArgument):

		status = http.StatusBadRequest

	case errors.Is(err, user.ErrInvalidCredentials):

		status = http.StatusUnauthorized
	default:

		slog.Error("internal server error", "err", err)
	}

	slog.Warn("client error", "status", status, "err", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
}
