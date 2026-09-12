package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/mapper"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase"
)

type Handler struct {
	service usecase.AppService
}

func NewHandler(service usecase.AppService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) NextTurn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameUUID, err := uuid.Parse(r.PathValue("uuid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if cg, err := h.service.GetGame(gameUUID); err == nil {
		if player, end := h.service.GameIsEnded(gameUUID); end {
			boardResponse := mapper.ToGameBoardResponse(cg, player, end)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(boardResponse)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var gameBoard dto.GameBoardRequest
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&gameBoard); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cg, err := h.service.ProcessPlayerMove(gameUUID, gameBoard.Board)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	winner, ended := h.service.GameIsEnded(gameUUID)
	boardResponse := mapper.ToGameBoardResponse(cg, winner, ended)

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(boardResponse)
}

func (h *Handler) NewGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cg, err := h.service.CreateGame()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := mapper.ToGameBoardResponse(cg, -1, false)

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameUUID, err := uuid.Parse(r.PathValue("uuid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, err := h.service.GetGame(gameUUID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(game.Board())
}
