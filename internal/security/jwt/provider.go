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
	"github.com/gr0shka/Tic-tac-toe/internal/domain/auth"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
)

type Claims struct {
	jwt.RegisteredClaims
	TokenType auth.TokenType `json:"token_type"`
}

func NewClaims(t *auth.Token, tt auth.TokenType) *Claims {
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        t.ID().String(),
			Subject:   t.UserID().String(),
			ExpiresAt: jwt.NewNumericDate(t.ExpiresAt()),
			IssuedAt:  jwt.NewNumericDate(t.IssuedAt()),
		},
		TokenType: tt,
	}
}

type provider struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewProvider(cfg config.JWT) (*provider, error) {
	privateText, err := os.ReadFile(cfg.PrivateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
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

func (p *provider) GenerateAccessToken(u user.User) (*auth.Token, error) {
	t := auth.NewToken(u.ID(), p.accessTTL)

	claims := NewClaims(t, auth.TokenTypeAccess)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to sign token: %w", err)
	}

	t.SetTokenString(tokenString)

	return t, nil
}

func (p *provider) GenerateRefreshToken(u user.User) (*auth.Token, error) {

	t := auth.NewToken(u.ID(), p.refreshTTL)

	claims := NewClaims(t, auth.TokenTypeRefresh)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to sign token: %w", err)
	}

	t.SetTokenString(tokenString)

	return t, nil
}

func (p *provider) ValidateAccessToken(at string) (*auth.Token, error) {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(at, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidToken, token.Header["alg"])
		}

		return p.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		return nil, fmt.Errorf("%w: unable to parse token: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("%w: invalid token", ErrInvalidToken)
	}

	if claims.TokenType != auth.TokenTypeAccess {
		return nil, fmt.Errorf("%w: invalid token type", ErrInvalidToken)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user id parse", ErrInvalidToken)
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid token id parse", ErrInvalidToken)
	}

	if claims.ExpiresAt == nil {
		return nil, fmt.Errorf("%w: invalid token expires at parse", ErrInvalidToken)
	}

	if claims.IssuedAt == nil {
		return nil, fmt.Errorf("%w: invalid token issued at parse", ErrInvalidToken)
	}

	t := auth.NewTokenWithFields(tokenID, userID, at, claims.ExpiresAt.Time, claims.IssuedAt.Time)

	return t, nil
}

func (p *provider) ValidateRefreshToken(rt string) (*auth.Token, error) {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(rt, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidToken, token.Header["alg"])
		}

		return p.publicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		return nil, fmt.Errorf("%w: unable to parse token: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("%w: invalid token", ErrInvalidToken)
	}

	if claims.ID == "" {
		return nil, fmt.Errorf("%w: invalid token id", ErrInvalidToken)
	}

	if claims.TokenType != auth.TokenTypeRefresh {
		return nil, fmt.Errorf("%w: invalid token type", ErrInvalidToken)
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid parse token id", ErrInvalidToken)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user id parse", ErrInvalidToken)
	}

	if claims.ExpiresAt == nil {
		return nil, fmt.Errorf("%w: invalid token expires at parse", ErrInvalidToken)
	}

	if claims.IssuedAt == nil {
		return nil, fmt.Errorf("%w: invalid token issued at parse", ErrInvalidToken)
	}

	t := auth.NewTokenWithFields(tokenID, userID, rt, claims.ExpiresAt.Time, claims.IssuedAt.Time)

	return t, nil
}

func (p *provider) GetUserIDFromToken(t string) (uuid.UUID, error) {

	token, err := jwt.Parse(t, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidToken, token.Header["alg"])
		}

		return p.publicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, ErrTokenExpired
		}

		return uuid.Nil, fmt.Errorf("%w: unable to parse token: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("%w: invalid token", ErrInvalidToken)
	}

	userIDStr, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: unable to get user ID", ErrInvalidToken)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: unable to parse user ID", ErrInvalidToken)
	}

	return userID, nil
}

func (p *provider) GetTokenIDFromToken(t string) (uuid.UUID, error) {

	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(t, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidToken, token.Header["alg"])
		}

		return p.publicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, ErrTokenExpired
		}

		return uuid.Nil, fmt.Errorf("%w: unable to parse token: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("%w: invalid token", ErrInvalidToken)
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: unable to parse token ID", ErrInvalidToken)
	}

	return tokenID, nil
}
