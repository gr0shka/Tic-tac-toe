package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/application/service"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/dto"
)

type Handler struct {
	service service.AppService
}

func NewHandler(service service.AppService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) NextTurn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	gameUUID, err := uuid.FromBytes([]byte(r.PathValue("uuid")))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
		return
	}

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(cg.GetBoard())
}

func (h *Handler) NewGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cg, err := h.service.CreateGame()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := dto.CreateGameResponse{
		Id:    cg.ID(),
		Board: cg.GetBoard(),
	}

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	gameUUID, err := uuid.FromBytes([]byte(r.PathValue("uuid")))
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
	encoder.Encode(game.GetBoard())
}
