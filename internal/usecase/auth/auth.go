package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
)

type AuthenticateService interface {
	Register(ctx context.Context, req dto.SignUpRequest) error
	Authenticate(ctx context.Context, login, password string) (*uuid.UUID, error)
}
