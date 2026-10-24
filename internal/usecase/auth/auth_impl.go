package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/security/jwt"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase"
	userService "github.com/gr0shka/Tic-tac-toe/internal/usecase/user"
	"golang.org/x/crypto/bcrypt"
)

type authenticateService struct {
	us   userService.UserService
	jwtP jwt.Provider
	repo Repository
	tx   usecase.Transactor
}

func NewAuthenticateService(us userService.UserService, jwtP jwt.Provider, repo Repository, tx usecase.Transactor) AuthenticateService {
	return &authenticateService{
		us:   us,
		jwtP: jwtP,
		repo: repo,
		tx:   tx,
	}
}

func (as *authenticateService) AuthenticateByToken(token string) (uuid.UUID, error) {
	t, err := as.jwtP.ValidateAccessToken(token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validate access token: %w", err)
	}

	if t.UserID() == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: validate access token invalid user id", jwt.ErrInvalidArgument)
	}

	return t.UserID(), nil
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

func (as *authenticateService) Authenticate(ctx context.Context, login, password string) (*JwtResponse, error) {
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

	accessToken, err := as.jwtP.GenerateAccessToken(*u)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := as.jwtP.GenerateRefreshToken(*u)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	err = as.repo.SaveRefreshToken(ctx, *refreshToken)
	if err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	response := &JwtResponse{
		Type:                  "Bearer",
		AccessToken:           accessToken.TokenString(),
		RefreshToken:          refreshToken.TokenString(),
		RefreshTokenExpiresAt: refreshToken.ExpiresAt(),
	}

	return response, nil
}

func (as *authenticateService) RefreshAccessToken(ctx context.Context, refreshToken string) (*JwtResponse, error) {
	token, err := as.jwtP.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("validate refresh token: %w", err)
	}

	_, err = as.repo.GetRefreshTokenFromID(ctx, token.ID())
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	u, err := as.us.GetUserByID(ctx, token.UserID())
	if err != nil {
		return nil, fmt.Errorf("get user by id from refresh token: %w", err)
	}

	newAccessToken, err := as.jwtP.GenerateAccessToken(*u)
	if err != nil {
		return nil, fmt.Errorf("generate new access token: %w", err)
	}

	response := &JwtResponse{
		Type:         "Bearer",
		AccessToken:  newAccessToken.TokenString(),
		RefreshToken: refreshToken,
	}

	return response, nil
}

func (as *authenticateService) RefreshRefreshToken(ctx context.Context, refreshToken string) (*JwtResponse, error) {
	token, err := as.jwtP.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("validate refresh token: %w", err)
	}

	u, err := as.us.GetUserByID(ctx, token.UserID())
	if err != nil {
		return nil, fmt.Errorf("get user by id from refresh token: %w", err)
	}

	var newRefreshToken *jwt.Token
	err = as.tx.Do(ctx, func(txCtx context.Context) error {

		err = as.repo.DeleteRefreshToken(txCtx, token.ID())
		if err != nil {
			return fmt.Errorf("delete refresh token: %w", err)
		}

		newRefreshToken, err = as.jwtP.GenerateRefreshToken(*u)
		if err != nil {
			return fmt.Errorf("generate new refresh token: %w", err)
		}

		err = as.repo.SaveRefreshToken(txCtx, *newRefreshToken)
		if err != nil {
			return fmt.Errorf("save new refresh token: %w", err)
		}

		return nil
	})
	if err != nil || newRefreshToken == nil {
		return nil, fmt.Errorf("update refresh token: %w", err)
	}

	newAccessToken, err := as.jwtP.GenerateAccessToken(*u)
	if err != nil {
		return nil, fmt.Errorf("generate new access token: %w", err)
	}

	response := &JwtResponse{
		Type:                  "Bearer",
		AccessToken:           newAccessToken.TokenString(),
		RefreshToken:          newRefreshToken.TokenString(),
		RefreshTokenExpiresAt: newRefreshToken.ExpiresAt(),
	}

	return response, nil
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
