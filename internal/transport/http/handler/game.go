package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase"
)

type CreateGameResponse struct {
	Id    uuid.UUID                               `json:"id"`
	Board [models.BoardSize][models.BoardSize]int `json:"board"`
}

type GameBoardRequest struct {
	Board [models.BoardSize][models.BoardSize]int `json:"board"`
}

type GameBoardResponse struct {
	GameIsEnded bool                                    `json:"game_is_ended"`
	Winner      int                                     `json:"winner"`
	Board       [models.BoardSize][models.BoardSize]int `json:"board"`
}

type Handler struct {
	service usecase.AppService
}

func NewHandler(service usecase.AppService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) NextTurn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	boardResponse := GameBoardResponse{}

	gameUUID, err := uuid.Parse(r.PathValue("uuid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if cg, err := h.service.GetGame(gameUUID); err == nil {
		if player, end := h.service.GameIsEnded(gameUUID); end {
			boardResponse.GameIsEnded = end
			boardResponse.Winner = player
			boardResponse.Board = cg.GetBoard()
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(boardResponse)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var gameBoard GameBoardRequest
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

	boardResponse.Board = cg.GetBoard()

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(boardResponse)
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

	data := CreateGameResponse{
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
	encoder.Encode(game.GetBoard())
}
