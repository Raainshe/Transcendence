-- +goose Up
CREATE TABLE files (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename   TEXT        NOT NULL,
    mime_type  TEXT        NOT NULL,
    size       BIGINT      NOT NULL,
    path       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE files;
