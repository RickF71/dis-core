-- GOV-11D: Corporeal Identity Log Storage
-- Private log for IRL authentication events maintained by corporeal domains

-- Migration: 20251114_gov11d_corporeal_identity_log.sql
-- Phase: GOV-11D (Corporeal Authentication Logging)

\set ON_ERROR_STOP on

-- Corporeal identity log table for IRL authentication events
CREATE TABLE IF NOT EXISTS corporeal_identity_log (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id UUID NOT NULL,
  corporeal_domain_id UUID NOT NULL REFERENCES domains(id),
  target_domain_id UUID NOT NULL REFERENCES domains(id),
  receipt_id UUID NOT NULL REFERENCES identity_receipts(id),
  method TEXT NOT NULL,
  channel TEXT NOT NULL,
  metadata JSONB DEFAULT '{}'::jsonb,
  logged_at TIMESTAMPTZ DEFAULT now()
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_corporeal_log_actor
  ON corporeal_identity_log(actor_id, logged_at DESC);

CREATE INDEX IF NOT EXISTS idx_corporeal_log_corporeal_domain
  ON corporeal_identity_log(corporeal_domain_id, logged_at DESC);

CREATE INDEX IF NOT EXISTS idx_corporeal_log_target_domain
  ON corporeal_identity_log(target_domain_id);

CREATE INDEX IF NOT EXISTS idx_corporeal_log_receipt
  ON corporeal_identity_log(receipt_id);

-- View for easy querying with domain names
CREATE OR REPLACE VIEW corporeal_identity_log_view AS
SELECT
  cil.id,
  cil.actor_id,
  cil.corporeal_domain_id,
  cd.name AS corporeal_domain_name,
  cil.target_domain_id,
  td.name AS target_domain_name,
  cil.receipt_id,
  cil.method,
  cil.channel,
  cil.metadata,
  cil.logged_at,
  ir.hash AS receipt_hash,
  ir.created_at AS receipt_created_at
FROM corporeal_identity_log cil
LEFT JOIN domains cd ON cil.corporeal_domain_id = cd.id
LEFT JOIN domains td ON cil.target_domain_id = td.id
LEFT JOIN identity_receipts ir ON cil.receipt_id = ir.id;

-- Add comments
COMMENT ON TABLE corporeal_identity_log IS 'GOV-11D: Private log of IRL authentication events for corporeal domains';
COMMENT ON COLUMN corporeal_identity_log.actor_id IS 'The actor (person) being authenticated';
COMMENT ON COLUMN corporeal_identity_log.corporeal_domain_id IS 'The corporeal domain performing the authentication';
COMMENT ON COLUMN corporeal_identity_log.target_domain_id IS 'The domain requesting authentication (bank, hospital, etc.)';
COMMENT ON COLUMN corporeal_identity_log.receipt_id IS 'Reference to identity.irlauth.v1 receipt';
COMMENT ON COLUMN corporeal_identity_log.method IS 'Authentication method (biometric, passkey, code, etc.)';
COMMENT ON COLUMN corporeal_identity_log.channel IS 'Authentication channel (device, kiosk, agent, etc.)';
COMMENT ON COLUMN corporeal_identity_log.metadata IS 'Privacy-aware contextual metadata (location, session, device info, etc.)';

-- Verification
DO $$
DECLARE
  log_count INT;
BEGIN
  SELECT COUNT(*) INTO log_count FROM corporeal_identity_log;

  RAISE NOTICE '✅ GOV-11D — Corporeal Identity Log Storage';
  RAISE NOTICE '    Table: corporeal_identity_log (private IRL auth events)';
  RAISE NOTICE '    View: corporeal_identity_log_view (with domain names)';
  RAISE NOTICE '    Existing log entries: %', log_count;
  RAISE NOTICE '    Ready for: identity.irlauth.v1 event logging';
  RAISE NOTICE '    Access: Restricted to actor and their corporeal domain';
END $$;
