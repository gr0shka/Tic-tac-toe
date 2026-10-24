package jwt

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidToken = errors.New("invalid token")
)
