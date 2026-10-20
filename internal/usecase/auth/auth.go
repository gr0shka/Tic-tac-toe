package auth

import (
	"context"
)

type AuthenticateService interface {
	Register(ctx context.Context, req SignUpRequest) error
	Authenticate(ctx context.Context, login, password string) (*JwtResponse, error)

	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	RefreshRefreshToken(ctx context.Context, refreshToken string) (string, error)
}
