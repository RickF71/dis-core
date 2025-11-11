-- sql/migrations/010_identity_bindings.sql

CREATE TABLE IF NOT EXISTS identity_bindings (
    id          SERIAL PRIMARY KEY,
    uid         TEXT NOT NULL,
    domain      TEXT NOT NULL,
    key         TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now(),
    UNIQUE (uid, domain)
);

