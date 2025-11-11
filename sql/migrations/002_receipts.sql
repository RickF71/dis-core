-- 002_receipts.sql
-- ci.call.v1 receipt log.

CREATE TABLE IF NOT EXISTS receipts (
    id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    type       TEXT NOT NULL,
    actor      TEXT,
    target     TEXT,
    domain     TEXT,
    payload    JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);
