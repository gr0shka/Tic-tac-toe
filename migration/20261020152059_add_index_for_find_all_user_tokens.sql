-- +goose Up
CREATE INDEX idx_jwt_token_user_id ON jwt_token(user_id);

-- +goose Down
DROP INDEX idx_jwt_token_user_id;
