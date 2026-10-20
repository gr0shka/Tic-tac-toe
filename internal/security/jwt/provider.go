package jwt

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/config"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
)

type Provider struct {
	repo       Repository
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewProvider(cfg config.JWT, repo Repository) (*Provider, error) {
	privateText, err := os.ReadFile(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read privet key: %w", err)
	}

	publicText, err := os.ReadFile(cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &Provider{
		repo:       repo,
		privateKey: privateKey,
		publicKey:  publicKey,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}, nil
}

type Token struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	IssuedAt  time.Time
}

func (p *Provider) GenerateAccessToken(u user.User) (string, error) {
	t := Token{
		ID:        uuid.New(),
		UserID:    u.ID(),
		ExpiresAt: time.Now().Add(p.accessTTL),
		IssuedAt:  time.Now(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.RegisteredClaims{
			ID:        t.ID.String(),
			Subject:   t.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(t.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(t.IssuedAt),
		})

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	return tokenString, nil
}

func (p *Provider) GenerateRefreshToken(ctx context.Context, u user.User) (string, error) {

	t := Token{
		ID:        uuid.New(),
		UserID:    u.ID(),
		ExpiresAt: time.Now().Add(p.refreshTTL),
		IssuedAt:  time.Now(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.RegisteredClaims{
			ID:        t.ID.String(),
			Subject:   t.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(t.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(t.IssuedAt),
		})

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	err = p.repo.SaveRefreshToken(ctx, t)
	if err != nil {
		return "", fmt.Errorf("unable to save refresh token: %w", err)
	}

	return tokenString, nil
}

func (p *Provider) ValidateAccessToken(at string) (bool, error) {

	token, err := jwt.Parse(at, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidArgument, token.Header["alg"])
		}

		return p.publicKey, nil
	})
	if err != nil {
		return false, fmt.Errorf("unable to parse token: %w", err)
	}

	if !token.Valid {
		return false, fmt.Errorf("%w: invalid token", ErrInvalidArgument)
	}

	date, err := token.Claims.GetExpirationTime()
	if err != nil {
		return false, fmt.Errorf("%w: invalid token claims", ErrInvalidArgument)
	}

	if date.After(time.Now()) {
		return false, fmt.Errorf("%w: access token expired", ErrTokenExpired)
	}

	return true, nil
}

func (p *Provider) ValidateRefreshToken(ctx context.Context, rt string) (bool, error) {

	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(rt, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidArgument, token.Header["alg"])
		}

		return p.publicKey, nil
	})
	if err != nil {
		return false, fmt.Errorf("unable to parse token: %w", err)
	}

	if !token.Valid {
		return false, fmt.Errorf("%w: invalid token", ErrInvalidArgument)
	}

	if claims.ID == "" {
		return false, fmt.Errorf("%w: invalid token id", ErrInvalidArgument)
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return false, fmt.Errorf("%w: invalid parse token id", ErrInvalidArgument)
	}

	repoToken, err := p.repo.GetRefreshTokenFromID(ctx, tokenID)
	if err != nil {
		return false, fmt.Errorf("unable to get refresh token: %w", err)
	}

	if repoToken.ExpiresAt.After(time.Now()) {
		return true, nil
	}

	return false, fmt.Errorf("%w: invalid token expired", ErrTokenExpired)
}

func (p *Provider) GetUserIDFromToken(t string) (uuid.UUID, error) {

	token, err := jwt.Parse(t, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidArgument, token.Header["alg"])
		}

		return p.publicKey, nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to parse token: %w", err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("%w: invalid token", ErrInvalidArgument)
	}

	userIDStr, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: unable to get user ID", ErrInvalidArgument)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: unable to parse user ID", ErrInvalidArgument)
	}

	return userID, nil
}

func (p *Provider) GetTokenIDFromToken(t string) (uuid.UUID, error) {

	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(t, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidArgument, token.Header["alg"])
		}

		return p.publicKey, nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to parse token: %w", err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("%w: invalid token", ErrInvalidArgument)
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: unable to parse token ID", ErrInvalidArgument)
	}

	return tokenID, nil
}
