package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/transactor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func (m *repository) AllGames(ctx context.Context) ([]*game.CurrentGame, error) {
	query := `SELECT * FROM current_game 
         WHERE player2_id IS NULL 
         LIMIT 50`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query all games: %w", err)
	}
	defer rows.Close()

	dto, err := pgx.CollectRows(rows, pgx.RowToStructByName[CurrentGameDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: no waiting games", game.ErrNotFound)
		}
		return nil, fmt.Errorf("all games collect rows: %w", err)
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

	query := `INSERT INTO current_game 
    (id,
     
     player1_id, 
     player1_real,
     player1_symbol,
     
     player2_id, 
     player2_real,
     player2_symbol,
     
     active_player_id,
     
     board, 
     number_of_turn, 
     status, 
     winner,
     winner_id,
     created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := m.db.Exec(ctx, query,
		dto.ID,

		dto.Player1ID,
		dto.Player1Real,
		dto.Player1Symbol,

		dto.Player2ID,
		dto.Player2Real,
		dto.Player2Symbol,

		dto.ActivePlayerID,

		dto.Board,
		dto.NumberOfTurn,
		dto.Status,
		dto.Winner,
		dto.WinnerID,
		dto.CreatedAt)
	if err != nil {
		return fmt.Errorf("save game with id = %s: %w", cg.ID(), err)
	}

	return nil
}

func (m *repository) Get(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	query := `SELECT * FROM current_game WHERE id = $1`

	db := transactor.GetExecutor(ctx, m.db)

	if transactor.IsTx(ctx) {
		query += ` FOR UPDATE`
	}

	rows, err := db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("get game with id = %s: %w", id, err)
	}

	dto, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[CurrentGameDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: no game with id = %s", game.ErrNotFound, id)
		}

		return nil, fmt.Errorf("get game collect rows: %w", err)
	}

	cg := DTOToDomain(dto)

	return cg, nil

}

func (m *repository) Update(ctx context.Context, cg *game.CurrentGame) error {
	dto := DomainToDTO(*cg)

	query := `UPDATE current_game
				SET 
				    player2_id = $2, 
				    player2_real = $3,
				    player2_symbol = $4,
				    
				    active_player_id = $5,
				    
				    board = $6, 
				    number_of_turn = $7, 
				    status = $8, 
				    winner = $9,
				    winner_id = $10
				WHERE id = $1`

	db := transactor.GetExecutor(ctx, m.db)

	_, err := db.Exec(ctx, query,
		dto.ID,

		dto.Player2ID,
		dto.Player2Real,
		dto.Player2Symbol,

		dto.ActivePlayerID,

		dto.Board,
		dto.NumberOfTurn,
		dto.Status,
		dto.Winner,
		dto.WinnerID)
	if err != nil {
		return fmt.Errorf("update game with id = %s: %w", cg.ID(), err)
	}

	return nil
}

func (m *repository) AllEndedGames(ctx context.Context, userID uuid.UUID) ([]*game.CurrentGame, error) {
	query := `SELECT * FROM current_game 
        WHERE (player1_id = $1 OR player2_id = $1) AND (status = $2 OR status = $3)
		ORDER BY created_at DESC`

	rows, err := m.db.Query(ctx, query, userID, game.StatusDraw, game.StatusPlayerWins)
	if err != nil {
		return nil, fmt.Errorf("get all ended games: %w", err)
	}
	defer rows.Close()

	dto, err := pgx.CollectRows(rows, pgx.RowToStructByName[CurrentGameDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: no waiting games", game.ErrNotFound)
		}

		return nil, fmt.Errorf("all games collect rows: %w", err)
	}

	result := make([]*game.CurrentGame, len(dto))
	for i, d := range dto {
		result[i] = DTOToDomain(d)
	}

	return result, nil
}
