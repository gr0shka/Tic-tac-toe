CREATE TABLE player (
    id UUID PRIMARY KEY,
    turn_of_number integer,
    real_player bool
);

CREATE TABLE current_game (
    id UUID PRIMARY KEY,
    player1_id UUID NOT NULL,
    player2_id UUID NOT NULL,
    board integer[][],
    number_of_turn integer,

    CONSTRAINT fk_player1
        FOREIGN KEY (player1_id)
        REFERENCES player(id),

    CONSTRAINT fk_player2
        FOREIGN KEY (player2_id)
        REFERENCES player(id)

);