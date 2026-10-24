package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/security/jwt"
)

type Repository interface {
	SaveRefreshToken(ctx context.Context, t jwt.Token) error
	GetRefreshTokenFromID(ctx context.Context, id uuid.UUID) (*jwt.Token, error)
	DeleteRefreshToken(ctx context.Context, id uuid.UUID) error
}
