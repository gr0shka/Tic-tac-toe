-- +goose Up
ALTER TABLE current_game ADD COLUMN winner_id UUID;

-- +goose Down
ALTER TABLE current_game DROP COLUMN winner_id;
