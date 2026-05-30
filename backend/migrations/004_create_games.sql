-- +goose Up
CREATE TABLE games (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    mode        VARCHAR(16) NOT NULL,
    status      VARCHAR(16) NOT NULL DEFAULT 'waiting',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE TABLE game_players (
    id            UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id       UUID    NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    user_id       UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score         INTEGER NOT NULL DEFAULT 0,
    lines_cleared INTEGER NOT NULL DEFAULT 0,
    level_reached INTEGER NOT NULL DEFAULT 1,
    placement     INTEGER,
    is_winner     BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(game_id, user_id)
);

-- +goose Down
DROP TABLE game_players;
DROP TABLE games;
