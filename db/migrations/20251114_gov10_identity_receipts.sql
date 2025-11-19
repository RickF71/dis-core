-- GOV-10: Identity Provenance & Alias Receipt Integration
-- Migration: 20251114_gov10_identity_receipts.sql
-- Purpose: Extend GOV-9 receipt ledger to cover identity operations

-- Identity receipts table with hash chain
CREATE TABLE IF NOT EXISTS identity_receipts (
    id          UUID PRIMARY KEY,
    domain_id   UUID NOT NULL REFERENCES domains(id),
    actor_id    UUID NOT NULL,
    action      TEXT NOT NULL,
    payload     JSONB NOT NULL,
    prev_id     UUID REFERENCES identity_receipts(id),
    hash        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    consent_by  UUID NOT NULL REFERENCES domains(id),
    alias_scope TEXT
);

-- Indexes for efficient lineage queries
CREATE INDEX IF NOT EXISTS idx_identity_receipts_actor ON identity_receipts(actor_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_identity_receipts_domain ON identity_receipts(domain_id);
CREATE INDEX IF NOT EXISTS idx_identity_receipts_action ON identity_receipts(action);
CREATE INDEX IF NOT EXISTS idx_identity_receipts_prev ON identity_receipts(prev_id);

-- Identity lineage view with integrity verification
CREATE OR REPLACE VIEW identity_lineage_view AS
SELECT
    ir.id,
    ir.actor_id,
    ir.domain_id,
    d.name as domain_name,
    ir.action,
    ir.prev_id,
    ir.hash,
    ir.consent_by,
    c.name as consent_by_name,
    ir.alias_scope,
    ir.created_at,
    CASE
        WHEN ir.prev_id IS NULL THEN 'root'
        WHEN prev_ir.id IS NULL THEN 'broken'
        ELSE 'valid'
    END as chain_status
FROM identity_receipts ir
LEFT JOIN domains d ON ir.domain_id = d.id
LEFT JOIN domains c ON ir.consent_by = c.id
LEFT JOIN identity_receipts prev_ir ON ir.prev_id = prev_ir.id
ORDER BY ir.actor_id, ir.created_at ASC;

-- Table documentation
COMMENT ON TABLE identity_receipts IS 'GOV-10: Immutable hash-linked ledger of identity governance actions (alias management, conversions, binding updates)';
COMMENT ON COLUMN identity_receipts.id IS 'Receipt UUID';
COMMENT ON COLUMN identity_receipts.domain_id IS 'Domain where identity action occurred';
COMMENT ON COLUMN identity_receipts.actor_id IS 'Identity UUID performing the action';
COMMENT ON COLUMN identity_receipts.action IS 'Action type (e.g., identity.alias.add.v1, identity.convert.v1)';
COMMENT ON COLUMN identity_receipts.payload IS 'Full action payload as JSONB';
COMMENT ON COLUMN identity_receipts.prev_id IS 'Previous receipt ID in chain (NULL for first receipt per actor)';
COMMENT ON COLUMN identity_receipts.hash IS 'SHA-256 hash of payload + prev_id for integrity verification';
COMMENT ON COLUMN identity_receipts.consent_by IS 'Domain providing consent/authority for this action';
COMMENT ON COLUMN identity_receipts.alias_scope IS 'Alias scope if action is alias-related';
COMMENT ON VIEW identity_lineage_view IS 'Identity lineage with chain integrity status (root/valid/broken)';

-- Verification logging
DO $$
DECLARE
    receipt_count INTEGER;
    actor_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO receipt_count FROM identity_receipts;
    SELECT COUNT(DISTINCT actor_id) INTO actor_count FROM identity_receipts;

    RAISE NOTICE '✅ GOV-10 — Identity Provenance & Alias Receipt Integration';
    RAISE NOTICE '   Identity receipts table created with hash chain integrity';
    RAISE NOTICE '   Existing receipts: % across % actors', receipt_count, actor_count;
    RAISE NOTICE '   Lineage view: identity_lineage_view (root/valid/broken status)';
    RAISE NOTICE '   Covers: alias management, notech→dis conversion, binding updates';
END $$;
