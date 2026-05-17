-- +goose Up
CREATE TABLE relationships (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       VARCHAR(16) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(requester_id, receiver_id)
);

-- +goose Down
DROP TABLE relationships;
