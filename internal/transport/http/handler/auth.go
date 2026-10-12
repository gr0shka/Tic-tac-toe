package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/auth"
)

type AuthHandler struct {
	service auth.AuthenticateService
}

func NewAuthHandler(service auth.AuthenticateService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dto.SignUpRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		responseError(w, fmt.Errorf("%w: error decoding request body", user.ErrInvalidArgument))
		return
	}

	authReq := auth.SignUpRequest{req.Login, req.Password}

	err = h.service.Register(r.Context(), authReq)
	if err != nil {
		responseError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered"))
}

func (h *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	login, pass, ok := r.BasicAuth()
	if !ok {
		responseError(w, fmt.Errorf("%w: basic auth required", user.ErrInvalidCredentials))
		return
	}

	userID, err := h.service.Authenticate(r.Context(), login, pass)
	if err != nil {
		responseError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(userID)
}
