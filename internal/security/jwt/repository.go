package jwt

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SaveRefreshToken(ctx context.Context, t Token) error
	GetRefreshTokenFromID(ctx context.Context, id uuid.UUID) (*Token, error)
}
