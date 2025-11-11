-- 001_policies.sql
-- Policy storage table.

CREATE TABLE IF NOT EXISTS policies (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    domain_id   TEXT,
    type        TEXT,
    rego_module TEXT,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ
);
