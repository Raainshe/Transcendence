-- +goose Up
CREATE TABLE lobbies (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    host_user_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_code   VARCHAR(8)  NOT NULL UNIQUE,
    max_players   SMALLINT    NOT NULL CHECK (max_players BETWEEN 2 AND 4),
    status        VARCHAR(16) NOT NULL DEFAULT 'waiting',
    game_id       UUID        REFERENCES games(id),
    shared_seed   BIGINT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lobbies_invite_code ON lobbies (invite_code);

CREATE TABLE lobby_members (
    id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    lobby_id  UUID        NOT NULL REFERENCES lobbies(id) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_ready  BOOLEAN     NOT NULL DEFAULT false,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (lobby_id, user_id)
);

CREATE INDEX idx_lobby_members_lobby_id ON lobby_members (lobby_id);

-- +goose Down
DROP TABLE lobby_members;
DROP TABLE lobbies;
