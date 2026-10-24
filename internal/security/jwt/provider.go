package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/config"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
)

type Provider interface {
	GenerateAccessToken(u user.User) (*Token, error)
	GenerateRefreshToken(u user.User) (*Token, error)
	ValidateAccessToken(at string) (*Token, error)
	ValidateRefreshToken(rt string) (*Token, error)
	GetUserIDFromToken(t string) (uuid.UUID, error)
	GetTokenIDFromToken(t string) (uuid.UUID, error)
}

type provider struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewProvider(cfg config.JWT) (Provider, error) {
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
		privateKey: privateKey,
		publicKey:  publicKey,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}, nil
}

func (p *provider) GenerateAccessToken(u user.User) (*Token, error) {
	t := NewToken(u.ID(), p.accessTTL)

	claims := NewClaims(t, TokenTypeAccess)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to sign token: %w", err)
	}

	t.SetTokenString(tokenString)

	return t, nil
}

func (p *provider) GenerateRefreshToken(u user.User) (*Token, error) {

	t := NewToken(u.ID(), p.refreshTTL)

	claims := NewClaims(t, TokenTypeRefresh)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to sign token: %w", err)
	}

	t.SetTokenString(tokenString)

	return t, nil
}

func (p *provider) ValidateAccessToken(at string) (*Token, error) {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(at, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidArgument, token.Header["alg"])
		}

		return p.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		return nil, fmt.Errorf("unable to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("%w: invalid token", ErrInvalidArgument)
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, fmt.Errorf("%w: invalid token type", ErrInvalidArgument)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user id parse", ErrInvalidArgument)
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid token id parse", ErrInvalidArgument)
	}

	t := NewTokenWithFields(tokenID, userID, at, claims.ExpiresAt.Time, claims.IssuedAt.Time)

	return t, nil
}

func (p *provider) ValidateRefreshToken(rt string) (*Token, error) {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(rt, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidArgument, token.Header["alg"])
		}

		return p.publicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		return nil, fmt.Errorf("unable to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("%w: invalid token", ErrInvalidArgument)
	}

	if claims.ID == "" {
		return nil, fmt.Errorf("%w: invalid token id", ErrInvalidArgument)
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, fmt.Errorf("%w: invalid token type", ErrInvalidArgument)
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid parse token id", ErrInvalidArgument)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user id parse", ErrInvalidArgument)
	}

	t := NewTokenWithFields(tokenID, userID, rt, claims.ExpiresAt.Time, claims.IssuedAt.Time)

	return t, nil
}

func (p *provider) GetUserIDFromToken(t string) (uuid.UUID, error) {

	token, err := jwt.Parse(t, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidArgument, token.Header["alg"])
		}

		return p.publicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, ErrTokenExpired
		}

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
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, ErrTokenExpired
		}

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
