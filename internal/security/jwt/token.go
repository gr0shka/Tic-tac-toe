package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	jwt.RegisteredClaims
	TokenType TokenType `json:"token_type"`
}

func NewClaims(t *Token, tt TokenType) *Claims {
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

type Token struct {
	id        uuid.UUID
	token     string
	userID    uuid.UUID
	expiresAt time.Time
	issuedAt  time.Time
}

func NewToken(userID uuid.UUID, expiresTTL time.Duration) *Token {
	return &Token{
		id:        uuid.New(),
		token:     "",
		userID:    userID,
		expiresAt: time.Now().Add(expiresTTL),
		issuedAt:  time.Now(),
	}
}

func NewTokenWithFields(tokenID, userID uuid.UUID, tokenStr string, expiresAt, issuedAt time.Time) *Token {
	return &Token{
		id:        tokenID,
		token:     tokenStr,
		userID:    userID,
		expiresAt: expiresAt,
		issuedAt:  issuedAt,
	}
}

func (t *Token) ID() uuid.UUID {
	return t.id
}

func (t *Token) TokenString() string {
	return t.token
}

func (t *Token) SetTokenString(token string) {
	t.token = token
}

func (t *Token) UserID() uuid.UUID {
	return t.userID
}

func (t *Token) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t *Token) IssuedAt() time.Time {
	return t.issuedAt
}
