-- +goose Up
CREATE INDEX idx_current_game_history_p1 ON current_game(player1_id, created_at DESC) WHERE status IN ('draw', 'player_wins');
CREATE INDEX idx_current_game_history_p2 ON current_game(player2_id, created_at DESC) WHERE status IN ('draw', 'player_wins');
-- +goose Down
DROP INDEX idx_current_game_history_p1
DROP INDEX idx_current_game_history_p2
