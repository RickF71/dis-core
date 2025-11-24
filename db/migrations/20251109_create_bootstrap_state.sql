-- Migration: create bootstrap_state table used to persist the bootstrap actor binding
CREATE TABLE IF NOT EXISTS bootstrap_state (
    actor_id TEXT
);

-- Ensure there is a single row to update/select against by default
INSERT INTO bootstrap_state (actor_id)
SELECT NULL
WHERE NOT EXISTS (SELECT 1 FROM bootstrap_state);

COMMENT ON TABLE bootstrap_state IS 'Holds the configured bootstrap actor binding (single-row table)';
