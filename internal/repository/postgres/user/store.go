package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
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
	query := `SELECT id, login FROM users WHERE id = $1`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDTO])
	if err != nil {
		return nil, err
	}

	return DTOtoDomain(user), nil
}

func (r repository) Save(ctx context.Context, user *user.User) error {
	query := "INSERT INTO users(id, login, password) VALUES($1, $2, $3)"

	u := DomainToDTO(user)

	_, err := r.db.Exec(ctx, query, u.ID, u.Login, u.Password)
	if err != nil {
		return err
	}

	return nil
}

func (r repository) Get(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := "SELECT * FROM users WHERE id = $1"

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDTO])
	if err != nil {
		return nil, err
	}

	return DTOtoDomain(u), nil
}

func (r repository) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	query := "SELECT * FROM users WHERE login = $1"

	rows, err := r.db.Query(ctx, query, login)
	if err != nil {
		return nil, err
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDTO])
	if err != nil {
		return nil, err
	}

	return DTOtoDomain(u), nil
}
