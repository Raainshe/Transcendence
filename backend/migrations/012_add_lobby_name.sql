-- +goose Up
ALTER TABLE lobbies ADD COLUMN name VARCHAR(64) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE lobbies DROP COLUMN name;