package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

	authReq := auth.SignUpRequest{
		Login:    req.Login,
		Password: req.Password,
	}

	err = h.service.Register(r.Context(), authReq)
	if err != nil {
		responseError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "successfully register"})
}

func (h *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dto.JwtRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		responseError(w, fmt.Errorf("%w: error decoding request body", user.ErrInvalidArgument))
		return
	}

	resp, err := h.service.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		responseError(w, err)
		return
	}

	response := dto.JwtResponse{
		Type:         resp.Type,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}

	setRefreshTokenInCookie(w, resp.RefreshToken, resp.RefreshTokenExpiresAt)

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	refreshToken, err := extractToken(r)
	if err != nil {
		responseError(w, err)
		return
	}

	resp, err := h.service.RefreshAccessToken(r.Context(), refreshToken)
	if err != nil {
		responseError(w, err)
		return
	}

	response := dto.JwtResponse{
		Type:         resp.Type,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) RefreshRefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	refreshToken, err := extractToken(r)
	if err != nil {
		responseError(w, err)
		return
	}

	resp, err := h.service.RefreshRefreshToken(r.Context(), refreshToken)
	if err != nil {
		responseError(w, err)
		return
	}

	response := dto.JwtResponse{
		Type:         resp.Type,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}

	setRefreshTokenInCookie(w, resp.RefreshToken, resp.RefreshTokenExpiresAt)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func setRefreshTokenInCookie(w http.ResponseWriter, refreshToken string, refreshTokenExpiresAt time.Time) {
	newCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  refreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, newCookie)
}

func extractToken(r *http.Request) (string, error) {

	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	if r.Body != nil {
		defer r.Body.Close()

		var request dto.RefreshJwtRequest
		err = json.NewDecoder(r.Body).Decode(&request)
		if err == nil || request.RefreshToken != "" {
			return request.RefreshToken, nil
		}
	}

	return "", fmt.Errorf("%w: refresh token required in cookie or body", user.ErrInvalidCredentials)
}
