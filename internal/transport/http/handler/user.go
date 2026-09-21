package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/user"
)

type UserHandler struct {
	service user.UserService
}

func NewUserHandler(service user.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.SignUpRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered"))
}

func (h *UserHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req := r.Header.Get("Authorization")

	u, err := h.service.Authenticate(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(u.ID())
}
