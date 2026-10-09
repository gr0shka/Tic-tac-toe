package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
)

type UserService interface {
	Create(ctx context.Context, login, passHash string) (*user.User, error)
	GetByLogin(ctx context.Context, login string) (*user.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}
