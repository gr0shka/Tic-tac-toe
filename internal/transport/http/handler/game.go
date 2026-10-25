package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
		responseError(w, fmt.Errorf(`%w: game uuid is not valid`, game.ErrInvalidArgument))
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		responseError(w, errors.New("user not in context"))
		return
	}

	var gameBoard dto.GameBoardRequest
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&gameBoard); err != nil {
		responseError(w, fmt.Errorf(`%w: invalid request body`, game.ErrInvalidArgument))
		return
	}

	nextCg, err := h.service.ProcessPlayerMove(r.Context(), userID, gameUUID, gameBoard.Board)
	if err != nil {
		responseError(w, err)
		return
	}

	gameBoardResponse := mapper.ToGameBoardResponse(nextCg)

	encoder := json.NewEncoder(w)
	encoder.Encode(gameBoardResponse)
}

func (h *GameHandler) NewGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		responseError(w, errors.New("user not in context"))
		return
	}

	var createRequest dto.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		responseError(w, fmt.Errorf("%w: error decoding create request", game.ErrInvalidArgument))
		return
	}

	cg, err := h.service.CreateGame(r.Context(), userID, game.GameMode(createRequest.Mode))
	if err != nil {
		responseError(w, err)
		return
	}

	data := mapper.ToGameBoardResponse(cg)

	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}

func (h *GameHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		responseError(w, errors.New("user not in context"))
		return
	}

	gameID, err := uuid.Parse(r.PathValue("uuid"))
	if err != nil {
		responseError(w, fmt.Errorf(`%w: game uuid is not valid`, game.ErrInvalidArgument))
		return
	}

	cg, err := h.service.JoinGame(r.Context(), gameID, userID)
	if err != nil {
		responseError(w, err)
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
		responseError(w, fmt.Errorf(`%w: game uuid is not valid`, game.ErrInvalidArgument))
		return
	}

	game, err := h.service.GetGame(r.Context(), gameUUID)
	if err != nil {
		responseError(w, err)
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
		responseError(w, err)
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

func (h *GameHandler) AllEndedGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		responseError(w, errors.New("user not in context"))
		return
	}

	endedGames, err := h.service.AllEndedGamesByUserID(r.Context(), userID)
	if err != nil {
		responseError(w, err)
		return
	}

	data := make([]dto.GameBoardResponse, len(endedGames))
	for i, g := range endedGames {
		data[i] = mapper.ToGameBoardResponse(g)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (h *GameHandler) GetLeaderBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limitInt := 10
	limit := r.URL.Query().Get("limit")

	if limit != "" {
		val, err := strconv.Atoi(limit)
		if err != nil {
			responseError(w, fmt.Errorf(`%w: invalid query parameter limit`, game.ErrInvalidArgument))
			return
		}
		limitInt = val
	}

	if limitInt > 100 {
		limitInt = 100
	}

	lb, err := h.service.GetLeaderBoard(r.Context(), limitInt)
	if err != nil {
		responseError(w, err)
		return
	}

	data := mapper.ToLeaderBoardResponse(lb)

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}
