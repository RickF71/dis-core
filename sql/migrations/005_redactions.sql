-- 005_redactions.sql
-- Redaction configuration for receipts / logs.

CREATE TABLE IF NOT EXISTS redactions (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    config     JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
