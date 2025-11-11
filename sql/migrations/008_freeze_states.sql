-- 008_freeze_states.sql
-- Domain freeze state persistence.

CREATE TABLE IF NOT EXISTS freeze_states (
    domain      TEXT PRIMARY KEY,
    active      BOOLEAN NOT NULL DEFAULT FALSE,
    reason      TEXT,
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);
