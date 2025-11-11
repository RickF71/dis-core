-- 007_identity_bindings.sql
-- Simple UID ↔ domain ↔ key bindings.

CREATE TABLE IF NOT EXISTS identity_bindings (
    id         SERIAL PRIMARY KEY,
    uid        TEXT NOT NULL,
    domain     TEXT NOT NULL,
    key        TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT identity_bindings_uid_domain_key UNIQUE (uid, domain)
);
