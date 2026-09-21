package user

import "github.com/google/uuid"

type DTO struct {
	ID       uuid.UUID `db:"id"`
	Login    string    `db:"login"`
	Password string    `db:"password"`
}
