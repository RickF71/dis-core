-- 013_freeze_persistence.sql
-- Finalize freeze_states with index support for active freezes and TTL.

CREATE INDEX IF NOT EXISTS idx_freeze_states_active ON freeze_states(active);
CREATE INDEX IF NOT EXISTS idx_freeze_states_expires_at ON freeze_states(expires_at);
