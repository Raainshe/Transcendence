-- +goose Up
CREATE TABLE oauth_providers (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         VARCHAR(32)  NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    access_token     TEXT,
    refresh_token    TEXT,
    expires_at       TIMESTAMPTZ,
    UNIQUE(provider, provider_user_id)
);

-- +goose Down
DROP TABLE oauth_providers;
