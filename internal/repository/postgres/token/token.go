package token

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/security/jwt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) jwt.Repository {
	return &tokenRepository{
		db: db,
	}
}

func (a tokenRepository) SaveRefreshToken(ctx context.Context, t jwt.Token) error {
	query := `INSERT INTO jwt_token (id, user_id, expires_at, issued_at) VALUES ($1, $2, $3, $4)`

	dto := tokenDTO{
		ID:        t.ID,
		UserID:    t.UserID,
		ExpiresAt: t.ExpiresAt,
		IssuedAt:  t.IssuedAt,
	}

	_, err := a.db.Exec(ctx, query, dto.ID, dto.UserID, dto.ExpiresAt, dto.IssuedAt)
	if err != nil {
		return fmt.Errorf("could not save refresh token: %w", err)
	}

	return nil
}

func (a tokenRepository) GetRefreshTokenFromID(ctx context.Context, id uuid.UUID) (*jwt.Token, error) {
	query := `SELECT id, user_id, expires_at, issued_at FROM jwt_token WHERE id = $1`

	rows, err := a.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("could not get refresh token: %w", err)
	}

	dto, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[tokenDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("could not get refresh token: %w", jwt.ErrNotFound)
		}

		return nil, fmt.Errorf("get refresh token from ID: %w", id)
	}

	token := &jwt.Token{
		ID:        dto.ID,
		UserID:    dto.UserID,
		ExpiresAt: dto.ExpiresAt,
		IssuedAt:  dto.IssuedAt,
	}

	return token, nil
}

func (a tokenRepository) DeleteRefreshToken(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM jwt_token WHERE id = $1`

	_, err := a.db.Exec(ctx, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: could not delete refresh token", jwt.ErrNotFound)
		}

		return fmt.Errorf("could not delete refresh token: %w", err)
	}

	return nil
}
