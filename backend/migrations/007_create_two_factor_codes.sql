-- +goose Up
CREATE TABLE two_factor_codes (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    purpose VARCHAR(16) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_two_factor_codes_user_purpose ON two_factor_codes (user_id, purpose);

-- +goose Down
DROP TABLE two_factor_codes;
