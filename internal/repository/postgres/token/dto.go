package token

import (
	"time"

	"github.com/google/uuid"
)

type tokenDTO struct {
	ID        uuid.UUID `db:"id"`
	Token     string    `db:"token"`
	UserID    uuid.UUID `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	IssuedAt  time.Time `db:"issued_at"`
}
