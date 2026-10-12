package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	userService "github.com/gr0shka/Tic-tac-toe/internal/usecase/user"
	"golang.org/x/crypto/bcrypt"
)

type authenticateService struct {
	us userService.UserService
}

func NewAuthenticateService(us userService.UserService) *authenticateService {
	return &authenticateService{
		us: us,
	}
}

func (as *authenticateService) Register(ctx context.Context, req SignUpRequest) error {
	if err := validateData(req.Login, req.Password); err != nil {
		return err
	}

	_, err := as.us.GetByLogin(ctx, req.Login)
	if err == nil {
		return user.ErrUserAlreadyExists
	}
	if !errors.Is(err, user.ErrUserNotFound) {
		return fmt.Errorf("check user existence: %w", err)
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = as.us.Create(ctx, req.Login, string(hashPass))
	if err != nil {
		return err
	}

	return nil
}

func (as *authenticateService) Authenticate(ctx context.Context, login, password string) (*uuid.UUID, error) {
	if err := validateData(login, password); err != nil {
		return nil, err
	}

	u, err := as.us.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.Password()), []byte(password)) != nil {
		return nil, user.ErrInvalidCredentials
	}

	return new(u.ID()), nil
}

func validateData(login, password string) error {
	if login == "" {
		return fmt.Errorf("%w: login cannot be empty", user.ErrInvalidArgument)
	}
	if password == "" {
		return fmt.Errorf("%w: password cannot be empty", user.ErrInvalidArgument)
	}

	return nil
}
