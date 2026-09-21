package user

import "github.com/google/uuid"

type User struct {
	id       uuid.UUID
	login    string
	password string
}

func New(id uuid.UUID, login string, password string) *User {
	return &User{
		id:       id,
		login:    login,
		password: password,
	}
}

func (u User) ID() uuid.UUID {
	return u.id
}

func (u User) Login() string {
	return u.login
}

func (u User) Password() string {
	return u.password
}
