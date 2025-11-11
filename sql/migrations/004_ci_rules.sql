-- 004_ci_rules.sql
-- CI / policy rules.

CREATE TABLE IF NOT EXISTS ci_rules (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    rules      JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
