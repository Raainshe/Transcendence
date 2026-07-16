
-- +goose Up
ALTER TABLE users
ADD COLUMN achievements JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE users DROP COLUMN achievements;


/* CREATE TABLE achievements (
    id
    user_id 
    title   VARCHAR(30),
    earned  BOOLEAN
) */