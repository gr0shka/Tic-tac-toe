package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/mapper"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/middleware"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/app"
)

type GameHandler struct {
	service app.AppService
}

func NewGameHandler(service app.AppService) *GameHandler {
	return &GameHandler{service: service}
}

func (h *GameHandler) NextTurn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameUUID, err := uuid.Parse(r.PathValue("uuid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDContextName).(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var gameBoard dto.GameBoardRequest
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&gameBoard); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nextCg, err := h.service.ProcessPlayerMove(r.Context(), userID, gameUUID, gameBoard.Board)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	gameBoardResponse := mapper.ToGameBoardResponse(nextCg)

	encoder := json.NewEncoder(w)
	if err = encoder.Encode(gameBoardResponse); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *GameHandler) NewGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDContextName).(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var createRequest dto.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cg, err := h.service.CreateGame(r.Context(), userID, game.GameMode(createRequest.Mode))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	data := mapper.ToGameBoardResponse(cg)

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}

func (h *GameHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDContextName).(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	gameID, err := uuid.Parse(r.PathValue("uuid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cg, err := h.service.JoinGame(r.Context(), gameID, userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	data := mapper.ToGameBoardResponse(cg)

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}

func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameUUID, err := uuid.Parse(r.PathValue("uuid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, err := h.service.GetGame(r.Context(), gameUUID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := mapper.ToGameBoardResponse(game)

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}

func (h *GameHandler) AllGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	games, err := h.service.AllGames(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := make([]dto.GameBoardResponse, len(games))
	for i, g := range games {
		data[i] = mapper.ToGameBoardResponse(g)
	}

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}
