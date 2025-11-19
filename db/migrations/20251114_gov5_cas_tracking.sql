-- GOV-5: CAS Version Tracking Migration
-- Adds Compare-And-Swap version tracking to identity_seats and authority_decisions tables

-- 1. Add version column to identity_seats for optimistic locking
ALTER TABLE identity_seats
  ADD COLUMN IF NOT EXISTS version BIGINT DEFAULT 1;

COMMENT ON COLUMN identity_seats.version IS
  'GOV-5: Version number for optimistic concurrency control (incremented on each update)';

-- Index for version queries
CREATE INDEX IF NOT EXISTS idx_identity_seats_version
  ON identity_seats(version);

-- 2. Add cas_ok to authority_decisions to track CAS success/failure explicitly
-- Note: The authority_decisions table already has cas_version_prev and cas_version_new columns
ALTER TABLE authority_decisions
  ADD COLUMN IF NOT EXISTS cas_ok BOOLEAN DEFAULT TRUE;

COMMENT ON COLUMN authority_decisions.cas_ok IS
  'GOV-5: Whether the CAS update succeeded (false indicates version conflict)';

-- Index for querying CAS conflicts
CREATE INDEX IF NOT EXISTS idx_authority_cas_ok
  ON authority_decisions(cas_ok) WHERE cas_ok = FALSE;

-- Index for CAS version range queries
CREATE INDEX IF NOT EXISTS idx_authority_cas_versions
  ON authority_decisions(cas_version_prev, cas_version_new);

-- View: CAS conflicts with enhanced details
CREATE OR REPLACE VIEW authority_decisions_cas_failures AS
SELECT
  id,
  domain_id,
  actor_id,
  seat,
  from_state,
  to_state,
  cas_version_prev,
  cas_version_new,
  allow,
  reason,
  created_at
FROM authority_decisions
WHERE cas_ok = FALSE
ORDER BY created_at DESC;

COMMENT ON VIEW authority_decisions_cas_failures IS
  'GOV-5: All CAS conflicts for race condition debugging and retry analysis';

-- Update existing records to have cas_ok=true (assume success if cas_applied=true)
UPDATE authority_decisions
SET cas_ok = cas_applied
WHERE cas_ok IS NULL;

-- Summary of migration
DO $$
BEGIN
  RAISE NOTICE '✅ GOV-5 CAS Tracking Migration Complete';
  RAISE NOTICE '   - Added version column to identity_seats';
  RAISE NOTICE '   - Created idx_identity_seats_version index';
  RAISE NOTICE '   - Added cas_ok column to authority_decisions';
  RAISE NOTICE '   - Created idx_authority_cas_ok index';
  RAISE NOTICE '   - Created idx_authority_cas_versions index';
  RAISE NOTICE '   - Created authority_decisions_cas_failures view';
  RAISE NOTICE '   - Backfilled cas_ok from cas_applied for existing records';
END $$;
