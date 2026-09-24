package middleware

import (
	"context"
	"net/http"

	"github.com/gr0shka/Tic-tac-toe/internal/usecase/user"
)

const UserIDContextName = "UserID"

type UserAuthenticator struct {
	service user.UserService
}

func NewUserAuthenticator(service user.UserService) *UserAuthenticator {
	return &UserAuthenticator{service: service}
}

func (ua *UserAuthenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r.Header.Get("Authorization")

		u, err := ua.service.Authenticate(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), UserIDContextName, u.ID()))
		next.ServeHTTP(w, r)
	})
}
