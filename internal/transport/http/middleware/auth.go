package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/auth"
)

type userIDContextKey struct{}

func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(uuid.UUID)
	return userID, ok
}

type UserAuthenticator struct {
	service auth.AuthenticateService
}

func NewUserAuthenticator(service auth.AuthenticateService) *UserAuthenticator {
	return &UserAuthenticator{service: service}
}

func (ua *UserAuthenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorizedJSON(w, "Missing Authorization header")
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorizedJSON(w, "Invalid Authorization header")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		uID, err := ua.service.AuthenticateByToken(tokenString)
		if err != nil {
			writeUnauthorizedJSON(w, "Invalid token")
			slog.Error("Failed to authenticate user", "error", err)
			return
		}

		r = r.WithContext(ContextWithUserID(r.Context(), uID))
		next.ServeHTTP(w, r)
	})
}

func writeUnauthorizedJSON(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
