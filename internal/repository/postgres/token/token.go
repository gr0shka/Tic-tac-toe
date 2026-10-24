package token

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/transactor"
	"github.com/gr0shka/Tic-tac-toe/internal/security/jwt"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) auth.Repository {
	return &tokenRepository{
		db: db,
	}
}

func (a tokenRepository) SaveRefreshToken(ctx context.Context, t jwt.Token) error {
	query := `INSERT INTO jwt_token (id ,user_id, expires_at, issued_at) VALUES ($1, $2, $3, $4)`

	dto := tokenDTO{
		ID:        t.ID(),
		UserID:    t.UserID(),
		ExpiresAt: t.ExpiresAt(),
		IssuedAt:  t.IssuedAt(),
	}

	db := transactor.GetExecutor(ctx, a.db)

	_, err := db.Exec(ctx, query, dto.ID, dto.UserID, dto.ExpiresAt, dto.IssuedAt)
	if err != nil {
		return fmt.Errorf("could not save refresh token: %w", err)
	}

	return nil
}

func (a tokenRepository) GetRefreshTokenFromID(ctx context.Context, id uuid.UUID) (*jwt.Token, error) {
	query := `SELECT id, user_id, expires_at, issued_at FROM jwt_token WHERE id = $1`

	db := transactor.GetExecutor(ctx, a.db)

	rows, err := db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("could not get refresh token: %w", err)
	}

	dto, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[tokenDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("could not get refresh token: %w", jwt.ErrNotFound)
		}

		return nil, fmt.Errorf("get refresh token from ID %v: %w", id, err)
	}

	token := jwt.NewTokenWithFields(dto.ID, dto.UserID, "", dto.ExpiresAt, dto.IssuedAt)

	return token, nil
}

func (a tokenRepository) DeleteRefreshToken(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM jwt_token WHERE id = $1`

	db := transactor.GetExecutor(ctx, a.db)

	tag, err := db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("could not delete refresh token: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: could not delete refresh token", jwt.ErrNotFound)
	}

	return nil
}
