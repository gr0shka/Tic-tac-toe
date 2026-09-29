package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo UserRepository
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func NewUserService(repo UserRepository) *userService {
	return &userService{repo: repo}
}

func (s *userService) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	if login == "" {
		return nil, user.ErrUserNotFound
	}

	u, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *userService) Register(ctx context.Context, login, password string) error {
	if err := validateData(login, password); err != nil {
		return err
	}

	if _, err := s.repo.GetByLogin(ctx, login); err == nil {
		return user.ErrUserAlreadyExists
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u := user.New(uuid.New(), login, string(hashPass))

	err = s.repo.Save(ctx, u)
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) Authenticate(ctx context.Context, login, password string) (*user.User, error) {
	if err := validateData(login, password); err != nil {
		return nil, err
	}

	u, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.Password()), []byte(password)) != nil {
		return nil, user.ErrPasswordNotMatch
	}

	return u, nil
}

func validateData(login, password string) error {
	if login == "" {
		return user.ErrInValidLogin
	}
	if password == "" {
		return user.ErrInValidPassword
	}

	return nil
}
