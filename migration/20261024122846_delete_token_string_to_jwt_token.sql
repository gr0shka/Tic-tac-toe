-- +goose Up
ALTER TABLE jwt_token DROP COLUMN token;

-- +goose Down
