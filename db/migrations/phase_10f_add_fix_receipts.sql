-- Phase 10F: Continuity Lineage Proofs - Add fix_receipts table
-- This table tracks rcpt.fix.v1 receipt types for policy continuity remediation lineage

CREATE TABLE IF NOT EXISTS fix_receipts (
    id TEXT PRIMARY KEY,
    original_receipt TEXT REFERENCES receipts(id) ON DELETE CASCADE,
    domain_ref TEXT NOT NULL,
    action_ref TEXT,
    policy_ref TEXT,
    fix_method TEXT NOT NULL,
    authorized_by TEXT NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    verification TEXT DEFAULT 'pending'
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_fix_original ON fix_receipts(original_receipt);
CREATE INDEX IF NOT EXISTS idx_fix_domain ON fix_receipts(domain_ref);
CREATE INDEX IF NOT EXISTS idx_fix_timestamp ON fix_receipts(timestamp);
CREATE INDEX IF NOT EXISTS idx_fix_verification ON fix_receipts(verification);

-- View for lineage proof tracking
CREATE OR REPLACE VIEW lineage_proofs AS
SELECT
    fr.id as fix_receipt_id,
    fr.original_receipt as original_receipt_id,
    r.event_id as original_event_id,
    r.policy_ref as original_policy_ref,
    fr.policy_ref as fixed_policy_ref,
    fr.fix_method,
    fr.authorized_by,
    fr.verification,
    fr.timestamp as fix_timestamp,
    r.issued_at as original_timestamp,
    CASE
        WHEN fr.verification = 'pending' THEN false
        ELSE true
    END as verified
FROM fix_receipts fr
LEFT JOIN receipts r ON fr.original_receipt = r.id
ORDER BY fr.timestamp DESC;

-- Comment for documentation
COMMENT ON TABLE fix_receipts IS 'Phase 10F: Tracks rcpt.fix.v1 receipts for policy continuity lineage proofs';
COMMENT ON VIEW lineage_proofs IS 'Phase 10F: Provides complete lineage proof tracking for continuity fixes';
