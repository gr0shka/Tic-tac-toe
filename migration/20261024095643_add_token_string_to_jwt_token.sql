-- +goose Up
ALTER TABLE jwt_token ADD COLUMN token TEXT NOT NULL UNIQUE;

-- +goose Down
ALTER TABLE jwt_token DROP COLUMN token;
