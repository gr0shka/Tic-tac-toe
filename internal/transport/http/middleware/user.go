package middleware

import (
	"net/http"

	"github.com/gr0shka/Tic-tac-toe/internal/usecase/user"
)

type UserAuthenticator struct {
	service user.UserService
}

func NewUserAuthenticator(service user.UserService) *UserAuthenticator {
	return &UserAuthenticator{service: service}
}

func (u *UserAuthenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r.Header.Get("Authorization")

		_, err := u.service.Authenticate(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
