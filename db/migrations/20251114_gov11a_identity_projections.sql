-- GOV-11A: Domain-Scoped Identity Projection & Corporeal Authentication
-- Extends GOV-10 identity_receipts to support domain projections and foreign identity acceptance

-- Migration: 20251114_gov11a_identity_projections.sql
-- Phase: GOV-11A (Data Model Extensions)

\set ON_ERROR_STOP on

-- Extend identity_receipts table with new columns for GOV-11
ALTER TABLE identity_receipts
  ADD COLUMN IF NOT EXISTS target_domain_id UUID REFERENCES domains(id),
  ADD COLUMN IF NOT EXISTS source_domain_id UUID REFERENCES domains(id),
  ADD COLUMN IF NOT EXISTS external_subject TEXT,
  ADD COLUMN IF NOT EXISTS channel TEXT,
  ADD COLUMN IF NOT EXISTS method TEXT,
  ADD COLUMN IF NOT EXISTS scope TEXT;

-- Add indexes for new query patterns
CREATE INDEX IF NOT EXISTS idx_identity_receipts_target_domain
  ON identity_receipts(target_domain_id) WHERE target_domain_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_identity_receipts_source_domain
  ON identity_receipts(source_domain_id) WHERE source_domain_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_identity_receipts_external_subject
  ON identity_receipts(external_subject) WHERE external_subject IS NOT NULL;

-- Update identity_lineage_view to include new fields
DROP VIEW IF EXISTS identity_lineage_view;
CREATE VIEW identity_lineage_view AS
SELECT
  ir.id,
  ir.actor_id,
  ir.domain_id,
  d.name AS domain_name,
  ir.action,
  ir.prev_id,
  ir.hash,
  ir.created_at,
  ir.consent_by,
  cb.name AS consent_by_domain_name,
  ir.alias_scope,
  ir.target_domain_id,
  td.name AS target_domain_name,
  ir.source_domain_id,
  sd.name AS source_domain_name,
  ir.external_subject,
  ir.channel,
  ir.method,
  ir.scope,
  CASE
    WHEN ir.prev_id IS NULL THEN 'root'
    WHEN NOT EXISTS (
      SELECT 1 FROM identity_receipts prev
      WHERE prev.id = ir.prev_id
    ) THEN 'broken'
    ELSE 'valid'
  END AS chain_status
FROM identity_receipts ir
LEFT JOIN domains d ON ir.domain_id = d.id
LEFT JOIN domains cb ON ir.consent_by = cb.id
LEFT JOIN domains td ON ir.target_domain_id = td.id
LEFT JOIN domains sd ON ir.source_domain_id = sd.id;

-- Add comments for new columns
COMMENT ON COLUMN identity_receipts.target_domain_id IS 'GOV-11: Target domain for domain-scoped identity operations (e.g., domain.id creation, IRL auth target)';
COMMENT ON COLUMN identity_receipts.source_domain_id IS 'GOV-11: Source domain for foreign identity acceptance (the external identity provider)';
COMMENT ON COLUMN identity_receipts.external_subject IS 'GOV-11: External identity subject/token from foreign domain (used in identity.accept.v1)';
COMMENT ON COLUMN identity_receipts.channel IS 'GOV-11: Authentication channel for IRL auth events (device, kiosk, agent, etc.)';
COMMENT ON COLUMN identity_receipts.method IS 'GOV-11: Authentication method for IRL auth events (biometric, passkey, code, etc.)';
COMMENT ON COLUMN identity_receipts.scope IS 'GOV-11: Scope for foreign identity acceptance (auth-only, auth+profile, etc.)';

-- Verification
DO $$
DECLARE
  receipt_count INT;
  new_columns_present BOOLEAN;
BEGIN
  -- Count existing receipts
  SELECT COUNT(*) INTO receipt_count FROM identity_receipts;

  -- Verify new columns exist
  SELECT COUNT(*) = 6 INTO new_columns_present
  FROM information_schema.columns
  WHERE table_name = 'identity_receipts'
  AND column_name IN ('target_domain_id', 'source_domain_id', 'external_subject', 'channel', 'method', 'scope');

  IF NOT new_columns_present THEN
    RAISE EXCEPTION 'GOV-11A migration failed: new columns not created';
  END IF;

  RAISE NOTICE '✅ GOV-11A — Identity Projection Schema Extensions';
  RAISE NOTICE '    New columns: target_domain_id, source_domain_id, external_subject, channel, method, scope';
  RAISE NOTICE '    Existing receipts preserved: %', receipt_count;
  RAISE NOTICE '    Updated identity_lineage_view with foreign identity fields';
  RAISE NOTICE '    Ready for: identity.domainid.*, identity.accept.*, identity.irlauth.* actions';
END $$;
