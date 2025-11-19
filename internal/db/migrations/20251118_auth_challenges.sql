-- +goose Up
CREATE TABLE IF NOT EXISTS auth_challenges (
    id UUID PRIMARY KEY,
    external_user_id UUID NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_auth_challenges_external_user_id
    ON auth_challenges (external_user_id);

-- +goose Down
DROP TABLE IF EXISTS auth_challenges;
