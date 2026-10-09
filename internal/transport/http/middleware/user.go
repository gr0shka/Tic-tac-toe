package middleware

import (
	"context"
	"net/http"

	"github.com/gr0shka/Tic-tac-toe/internal/usecase/auth"
)

const UserIDContextName = "UserID"

type UserAuthenticator struct {
	service auth.AuthenticateService
}

func NewUserAuthenticator(service auth.AuthenticateService) *UserAuthenticator {
	return &UserAuthenticator{service: service}
}

func (ua *UserAuthenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login, pass, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		uID, err := ua.service.Authenticate(r.Context(), login, pass)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), UserIDContextName, uID.String()))
		next.ServeHTTP(w, r)
	})
}
