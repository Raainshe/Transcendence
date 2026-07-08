-- +goose Up
CREATE TABLE messages (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    read_at TIMESTAMPTZ
);
CREATE INDEX idx_messages_pair_created ON messages (sender_id, recipient_id, created_at);
CREATE INDEX idx_messages_unread ON messages (recipient_id, read_at);

-- +goose Down
DROP TABLE messages;
