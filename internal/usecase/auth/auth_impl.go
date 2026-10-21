package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/security/jwt"
	userService "github.com/gr0shka/Tic-tac-toe/internal/usecase/user"
	"golang.org/x/crypto/bcrypt"
)

type authenticateService struct {
	us   userService.UserService
	jwtP *jwt.Provider
}

func (as *authenticateService) AuthenticateByToken(token string) (uuid.UUID, error) {
	userID, err := as.jwtP.GetUserIDFromToken(token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get user ID from token: %w", err)
	}

	return userID, nil
}

func NewAuthenticateService(us userService.UserService, jwtP *jwt.Provider) *authenticateService {
	return &authenticateService{
		us:   us,
		jwtP: jwtP,
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

	refreshToken, err := as.jwtP.GenerateRefreshToken(ctx, *u)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	response := &JwtResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return response, nil
}

func (as *authenticateService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	if ok, err := as.jwtP.ValidateRefreshToken(ctx, refreshToken); !ok || err != nil {
		return "", fmt.Errorf("validate refresh token: %w", err)
	}

	userID, err := as.jwtP.GetUserIDFromToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("get user id from refresh token: %w", err)
	}

	u, err := as.us.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get user by id from refresh token: %w", err)
	}

	newAccessToken, err := as.jwtP.GenerateAccessToken(*u)
	if err != nil {
		return "", fmt.Errorf("generate new access token: %w", err)
	}

	return newAccessToken, nil
}

func (as *authenticateService) RefreshRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	if ok, err := as.jwtP.ValidateRefreshToken(ctx, refreshToken); !ok || err != nil {
		return "", fmt.Errorf("validate refresh token: %w", err)
	}

	tokenID, err := as.jwtP.GetTokenIDFromToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("get token id from refresh token: %w", err)
	}

	userID, err := as.jwtP.GetUserIDFromToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("get user id from refresh token: %w", err)
	}

	u, err := as.us.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get user by id from refresh token: %w", err)
	}

	newToken, err := as.jwtP.UpdateRefreshTokenFromID(ctx, tokenID, u)
	if err != nil {
		return "", fmt.Errorf("update refresh token: %w", err)
	}

	return newToken, nil
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
