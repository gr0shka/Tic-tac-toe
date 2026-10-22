-- +goose Up
ALTER TABLE current_game ADD COLUMN created_at timestamptz NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE current_game DROP COLUMN created_at;
