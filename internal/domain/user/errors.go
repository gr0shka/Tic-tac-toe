package user

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrPasswordNotMatch  = errors.New("password does not match")
	ErrLoginNotMatch     = errors.New("login does not match")
	ErrInValidPassword   = errors.New("invalid password")
	ErrInValidLogin      = errors.New("invalid login")
)
