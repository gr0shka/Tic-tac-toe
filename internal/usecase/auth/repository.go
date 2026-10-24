package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/auth"
)

type Repository interface {
	SaveRefreshToken(ctx context.Context, t auth.Token) error
	GetRefreshTokenFromID(ctx context.Context, id uuid.UUID) (*auth.Token, error)
	DeleteRefreshToken(ctx context.Context, id uuid.UUID) error
}
