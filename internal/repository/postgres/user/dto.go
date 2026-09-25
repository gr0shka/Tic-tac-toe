package user

import "github.com/google/uuid"

type userDTO struct {
	ID       uuid.UUID `db:"id"`
	Login    string    `db:"login"`
	Password string    `db:"password"`
}
