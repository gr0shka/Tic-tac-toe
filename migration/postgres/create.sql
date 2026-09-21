CREATE TABLE current_game (
    id UUID PRIMARY KEY,

    player1_id UUID NOT NULL,
    player1_real bool,

    player2_id UUID NOT NULL,
    player2_real bool,

    board integer[],
    number_of_turn integer,
    is_ended bool,
    winner integer
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    login varchar(32) unique NOT NULL,
    password varchar(32) NOT NULL
);