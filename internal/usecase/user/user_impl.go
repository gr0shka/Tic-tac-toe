package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
)

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *userService {
	return &userService{repo: repo}
}

func (s *userService) Create(ctx context.Context, login, passHash string) (*user.User, error) {
	u := user.New(uuid.New(), login, passHash)

	err := s.repo.Save(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *userService) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	if login == "" {
		return nil, fmt.Errorf("%w: login cannot be empty", user.ErrInvalidArgument)
	}

	u, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	return u, nil
}
