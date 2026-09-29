-- +goose Up
CREATE TABLE current_game (
    id UUID PRIMARY KEY,

    player1_id UUID NOT NULL,
    player1_symbol integer,
    player1_real bool,

    player2_id UUID,
    player2_symbol integer,
    player2_real bool,

    active_player_id UUID,

    board integer[],
    number_of_turn integer,
    status varchar(32),
    winner integer
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    login varchar(32) unique NOT NULL,
    password varchar(32) NOT NULL
);

-- +goose Down
DROP TABLE current_game;
DROP TABLE users;
