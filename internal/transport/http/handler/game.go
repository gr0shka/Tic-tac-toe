package handler

import (
	"context"
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

	var gameBoard dto.GameBoardRequest
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&gameBoard); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nextCg, err := h.service.ProcessPlayerMove(context.Background(), gameUUID, gameBoard.Board)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	gameBoardResponse := mapper.ToGameBoardResponse(nextCg)

	encoder := json.NewEncoder(w)
	if err = encoder.Encode(gameBoardResponse); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *Handler) NewGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cg, err := h.service.CreateGame(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := mapper.ToGameBoardResponse(cg)

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

	game, err := h.service.GetGame(context.Background(), gameUUID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := mapper.ToGameBoardResponse(game)

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}
