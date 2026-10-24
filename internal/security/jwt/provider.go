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

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Provider interface {
	GenerateAccessToken(u user.User) (string, error)
	GenerateRefreshToken(ctx context.Context, u user.User) (string, time.Time, error)
	ValidateAccessToken(at string) (uuid.UUID, error)
	ValidateRefreshToken(ctx context.Context, rt string) (bool, error)
	GetUserIDFromToken(t string) (uuid.UUID, error)
	GetTokenIDFromToken(t string) (uuid.UUID, error)
	UpdateRefreshTokenFromID(ctx context.Context, tokenID uuid.UUID, u *user.User) (string, time.Time, error)
}

type Claims struct {
	jwt.RegisteredClaims
	TokenType TokenType `json:"token_type"`
}

type provider struct {
	repo       Repository
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewProvider(cfg config.JWT, repo Repository) (Provider, error) {
	privateText, err := os.ReadFile(cfg.PrivateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read privet key: %w", err)
	}

	publicText, err := os.ReadFile(cfg.PublicFile)
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

	return &provider{
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

func (p *provider) GenerateAccessToken(u user.User) (string, error) {
	t := Token{
		ID:        uuid.New(),
		UserID:    u.ID(),
		ExpiresAt: time.Now().Add(p.accessTTL),
		IssuedAt:  time.Now(),
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        t.ID.String(),
			Subject:   t.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(t.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(t.IssuedAt),
		},
		TokenType: TokenTypeAccess,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	return tokenString, nil
}

func (p *provider) GenerateRefreshToken(ctx context.Context, u user.User) (string, time.Time, error) {

	t := Token{
		ID:        uuid.New(),
		UserID:    u.ID(),
		ExpiresAt: time.Now().Add(p.refreshTTL),
		IssuedAt:  time.Now(),
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        t.ID.String(),
			Subject:   t.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(t.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(t.IssuedAt),
		},
		TokenType: TokenTypeRefresh,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("unable to sign token: %w", err)
	}

	err = p.repo.SaveRefreshToken(ctx, t)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("unable to save refresh token: %w", err)
	}

	return tokenString, t.ExpiresAt, nil
}

func (p *provider) ValidateAccessToken(at string) (uuid.UUID, error) {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(at, claims, func(token *jwt.Token) (any, error) {
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

	if claims.TokenType != TokenTypeAccess {
		return uuid.Nil, fmt.Errorf("%w: invalid token type", ErrInvalidArgument)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid token id parse", ErrInvalidArgument)
	}

	return userID, nil
}

func (p *provider) ValidateRefreshToken(ctx context.Context, rt string) (bool, error) {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(rt, claims, func(token *jwt.Token) (any, error) {
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

	if claims.TokenType != TokenTypeRefresh {
		return false, fmt.Errorf("%w: invalid token type", ErrInvalidArgument)
	}

	if repoToken.ExpiresAt.After(time.Now()) {
		return true, nil
	}

	return false, fmt.Errorf("%w: invalid token expired", ErrTokenExpired)
}

func (p *provider) GetUserIDFromToken(t string) (uuid.UUID, error) {

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

func (p *provider) GetTokenIDFromToken(t string) (uuid.UUID, error) {

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

func (p *provider) UpdateRefreshTokenFromID(ctx context.Context, tokenID uuid.UUID, u *user.User) (string, time.Time, error) {
	if tokenID == uuid.Nil {
		return "", time.Time{}, fmt.Errorf("%w: invalid token id", ErrInvalidArgument)
	}
	if u == nil {
		return "", time.Time{}, fmt.Errorf("%w: invalid user", ErrInvalidArgument)
	}

	if err := p.repo.DeleteRefreshToken(ctx, tokenID); err != nil {
		return "", time.Time{}, fmt.Errorf("unable to delete refresh token: %w", err)
	}

	newRefreshToken, refreshTTL, err := p.GenerateRefreshToken(ctx, *u)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("unable to generate refresh token: %w", err)
	}

	return newRefreshToken, refreshTTL, nil
}
