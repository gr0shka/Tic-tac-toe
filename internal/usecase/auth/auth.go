package auth

import (
	"context"

	"github.com/google/uuid"
)

type AuthenticateService interface {
	Register(ctx context.Context, req SignUpRequest) error
	Authenticate(ctx context.Context, login, password string) (*uuid.UUID, error)
}
