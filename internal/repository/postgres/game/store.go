package game

import (
	"context"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func (m *repository) AllGames(ctx context.Context) ([]*game.CurrentGame, error) {
	query := `SELECT * FROM current_game WHERE player2_id IS NULL`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, game.ErrNotFound
	}
	defer rows.Close()

	dto, err := pgx.CollectRows(rows, pgx.RowToStructByName[CurrentGameDTO])
	if err != nil {
		return nil, game.ErrNotFound
	}

	result := make([]*game.CurrentGame, len(dto))
	for i, d := range dto {
		result[i] = DTOToDomain(d)
	}

	return result, nil
}

func New(db *pgxpool.Pool) *repository {
	return &repository{
		db: db,
	}
}

func (m *repository) Save(ctx context.Context, cg *game.CurrentGame) error {
	dto := DomainToDTO(*cg)

	query := `INSERT INTO current_game (id, player1_id, player1_real, player2_id, player2_real, board, number_of_turn, status, winner)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := m.db.Exec(ctx, query, dto.ID, dto.Player1ID, dto.Player1Real, dto.Player2ID, dto.Player2Real, dto.Board, dto.NumberOfTurn, dto.Status, dto.Winner)
	if err != nil {
		return err
	}

	return nil
}

func (m *repository) Get(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	rows, err := m.db.Query(ctx, "SELECT * FROM current_game WHERE id=$1", id)
	if err != nil {
		return nil, game.ErrNotFound
	}

	dto, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[CurrentGameDTO])
	if err != nil {
		return nil, game.ErrNotFound
	}

	cg := DTOToDomain(dto)

	return cg, nil
}

func (m *repository) Update(ctx context.Context, cg *game.CurrentGame) error {
	dto := DomainToDTO(*cg)

	query := `UPDATE current_game
				SET player2_id = $2, player2_real = $3, board = $4, number_of_turn = $5, status = $6, winner = $7
				WHERE id = $1`

	_, err := m.db.Exec(ctx, query, dto.ID, dto.Player2ID, dto.Player2Real, dto.Board, dto.NumberOfTurn, dto.Status, dto.Winner)
	if err != nil {
		return err
	}

	return nil
}
