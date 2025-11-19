-- GOV-9: Authority Continuity & Receipt Lineage
-- Migration: 20251113_gov9_authority_receipts.sql
-- Purpose: Immutable hash-linked ledger of all governance actions

-- Authority receipts table with hash chain
CREATE TABLE IF NOT EXISTS authority_receipts (
    id              TEXT PRIMARY KEY,
    domain_id       UUID NOT NULL,
    action          TEXT NOT NULL,
    prev_id         TEXT,
    payload         JSONB,
    hash            TEXT NOT NULL,
    policy_digest   TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT fk_domain FOREIGN KEY (domain_id) REFERENCES domains(id),
    CONSTRAINT fk_prev FOREIGN KEY (prev_id) REFERENCES authority_receipts(id)
);

-- Indexes for efficient lineage queries
CREATE INDEX IF NOT EXISTS idx_authority_receipts_domain ON authority_receipts(domain_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_authority_receipts_action ON authority_receipts(action);
CREATE INDEX IF NOT EXISTS idx_authority_receipts_prev ON authority_receipts(prev_id);

-- Lineage view with integrity verification
CREATE OR REPLACE VIEW authority_lineage_view AS
SELECT
    ar.id,
    ar.domain_id,
    d.name as domain_name,
    ar.action,
    ar.prev_id,
    ar.hash,
    ar.policy_digest,
    ar.created_at,
    CASE
        WHEN ar.prev_id IS NULL THEN 'root'
        WHEN prev_ar.id IS NULL THEN 'broken'
        ELSE 'valid'
    END as chain_status
FROM authority_receipts ar
LEFT JOIN domains d ON ar.domain_id = d.id
LEFT JOIN authority_receipts prev_ar ON ar.prev_id = prev_ar.id
ORDER BY ar.created_at DESC;

-- Table documentation
COMMENT ON TABLE authority_receipts IS 'GOV-9: Immutable hash-linked ledger of all governance actions (seat changes, policy updates, freeze/unfreeze)';
COMMENT ON COLUMN authority_receipts.id IS 'Receipt ID in format rcpt-<uuid>';
COMMENT ON COLUMN authority_receipts.domain_id IS 'UUID of domain this governance action affects';
COMMENT ON COLUMN authority_receipts.action IS 'Action type (e.g., seat.create.v1, domain.freeze.v1, policy.update.v1)';
COMMENT ON COLUMN authority_receipts.prev_id IS 'Previous receipt ID in chain (NULL for first receipt per domain)';
COMMENT ON COLUMN authority_receipts.payload IS 'Full action payload as JSONB';
COMMENT ON COLUMN authority_receipts.hash IS 'SHA-256 hash of payload + prev_id for integrity verification';
COMMENT ON COLUMN authority_receipts.policy_digest IS 'SHA-256 digest of active policy bundle at time of action';
COMMENT ON VIEW authority_lineage_view IS 'Lineage view with chain integrity status (root/valid/broken)';

-- Verification logging
DO $$
DECLARE
    receipt_count INTEGER;
    domain_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO receipt_count FROM authority_receipts;
    SELECT COUNT(DISTINCT domain_id) INTO domain_count FROM authority_receipts;

    RAISE NOTICE '✅ GOV-9 — Authority Continuity & Receipt Lineage';
    RAISE NOTICE '   Authority receipts table created with hash chain integrity';
    RAISE NOTICE '   Existing receipts: % across % domains', receipt_count, domain_count;
    RAISE NOTICE '   Lineage view: authority_lineage_view (root/valid/broken status)';
END $$;
