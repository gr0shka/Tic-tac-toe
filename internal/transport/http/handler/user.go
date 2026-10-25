package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	userDomain "github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/middleware"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/user"
)

type UserHandler struct {
	service user.UserService
}

func NewUserHandler(service user.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := uuid.Parse(r.PathValue("uuid"))
	if err != nil {
		responseError(w, fmt.Errorf(`%w: user uuid is not valid`, userDomain.ErrInvalidArgument))
		return
	}

	u, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		responseError(w, err)
		return
	}

	resp := dto.UserInfoResponse{
		ID:    u.ID(),
		Login: u.Login(),
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetUserByToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		responseError(w, errors.New("user not in context"))
		return
	}

	u, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		responseError(w, err)
		return
	}

	resp := dto.UserInfoResponse{
		ID:    u.ID(),
		Login: u.Login(),
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
}
