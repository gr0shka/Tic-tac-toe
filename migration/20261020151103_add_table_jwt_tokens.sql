-- +goose Up
CREATE TABLE jwt_token (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE jwt_token;
