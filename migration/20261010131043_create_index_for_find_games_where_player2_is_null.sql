-- +goose Up
CREATE INDEX ind_current_game_waiting ON current_game(player2_id) WHERE player2_id IS NULL;

-- +goose Down
DROP INDEX ind_current_game_waiting;
