-- Migration: Add domain freeze state table
-- File: db/migrations/20251120_add_domain_freeze_state.sql

CREATE TABLE IF NOT EXISTS domain_freeze_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    reason TEXT NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ttl_until TIMESTAMPTZ NULL,
    override_of UUID NULL REFERENCES domain_freeze_state(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Unique active freeze per domain+scope
CREATE UNIQUE INDEX IF NOT EXISTS idx_domain_freeze_unique_active ON domain_freeze_state(domain_id, scope) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_domain_freeze_domain_active ON domain_freeze_state(domain_id, is_active);

COMMENT ON TABLE domain_freeze_state IS 'Per-domain freeze state records for authority-driven freezes (freeze/unfreeze/override)';
COMMENT ON COLUMN domain_freeze_state.scope IS 'Scope of the freeze (e.g., all, policy, write)';
COMMENT ON COLUMN domain_freeze_state.reason IS 'Human readable reason for the freeze';
