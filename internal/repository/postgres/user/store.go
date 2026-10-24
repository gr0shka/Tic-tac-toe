package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/transactor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *repository {
	return &repository{
		db: db,
	}
}

func (r repository) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `
			SELECT id, login, password 
			FROM users 
			WHERE id = $1`

	db := transactor.GetExecutor(ctx, r.db)

	rows, err := db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("get user with id %v: %w", id, err)
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: get user with id %v", user.ErrUserNotFound, id)
		}

		return nil, err
	}

	return DTOtoDomain(u), nil
}

func (r repository) Save(ctx context.Context, u *user.User) error {
	query := `INSERT INTO users(id, login, password) 
				VALUES($1, $2, $3)
				ON CONFLICT(login) DO NOTHING`

	uDTO := DomainToDTO(u)

	db := transactor.GetExecutor(ctx, r.db)

	tag, err := db.Exec(ctx, query, uDTO.ID, uDTO.Login, uDTO.Password)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: user already exists", user.ErrUserAlreadyExists)
	}

	return nil
}

func (r repository) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	query := `
			SELECT id, login, password 
			FROM users 
			WHERE login = $1`

	db := transactor.GetExecutor(ctx, r.db)

	rows, err := db.Query(ctx, query, login)
	if err != nil {
		return nil, err
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: get user with login %v", user.ErrUserNotFound, login)
		}
		return nil, err
	}

	return DTOtoDomain(u), nil
}
