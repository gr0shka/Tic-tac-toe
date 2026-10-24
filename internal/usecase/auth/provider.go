package auth

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/auth"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
)

type Provider interface {
	GenerateAccessToken(u user.User) (*auth.Token, error)
	GenerateRefreshToken(u user.User) (*auth.Token, error)
	ValidateAccessToken(at string) (*auth.Token, error)
	ValidateRefreshToken(rt string) (*auth.Token, error)
	GetUserIDFromToken(t string) (uuid.UUID, error)
	GetTokenIDFromToken(t string) (uuid.UUID, error)
}
