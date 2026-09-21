package user

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
)

type userService struct {
	repo UserRepository
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

	u := user.New(uuid.New(), login, password)

	err := s.repo.Save(ctx, u)
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) Authenticate(ctx context.Context, auth string) (*user.User, error) {
	login, password := reBase64(auth)

	if err := validateData(login, password); err != nil {
		return nil, err
	}

	u, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	if u.Password() != password {
		return nil, user.ErrPasswordNotMatch
	}

	return u, nil
}

func reBase64(s string) (login, password string) {
	s = strings.TrimPrefix(s, "Basic ")

	str, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return
	}

	args := strings.Split(string(str), ":")

	if len(args) != 2 {
		return "", ""
	}

	return args[0], args[1]
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
