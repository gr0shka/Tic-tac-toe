package jwt

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrTokenExpired    = errors.New("token expired")
)
