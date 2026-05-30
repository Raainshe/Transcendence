-- +goose Up
CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username       VARCHAR(32)  UNIQUE NOT NULL,
    email          VARCHAR(255) UNIQUE NOT NULL,
    password_hash  TEXT,
    avatar_url     TEXT,
    role           VARCHAR(16)  NOT NULL DEFAULT 'user',
    is_2fa_enabled BOOLEAN      NOT NULL DEFAULT false,
    totp_secret    TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    last_seen_at   TIMESTAMPTZ
);

-- +goose Down
DROP TABLE users;
