-- 003_thresholds.sql
-- Threshold configuration.

CREATE TABLE IF NOT EXISTS thresholds (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    config     JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
