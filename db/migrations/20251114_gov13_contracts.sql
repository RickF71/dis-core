-- GOV-13: Contracts Table & DSCI Contract Wiring
-- Migration: 20251114_gov13_contracts.sql
-- Purpose: Create canonical contracts table for DSCI-backed domain relationships

-- ============================================================
-- CONTRACTS TABLE
-- ============================================================
-- Stores DSCI (Domain-Signed Contract Inheritance) contracts that govern
-- domain participation, membership, data processing, and other consent relationships.
--
-- Key concepts:
-- - domain_id: The issuing domain (contract owner)
-- - subject_domain_id: The domain subject to the contract (usually a RELATIONSHIP alias)
-- - alias_id: Optional specific alias used when accepting the contract
-- - DSCI triple: (channel, reference, version) uniquely identifies the contract artifact

CREATE TABLE IF NOT EXISTS contracts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id           UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    subject_domain_id   UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    alias_id            UUID NULL REFERENCES domains(id) ON DELETE SET NULL,
    contract_type       TEXT NOT NULL,
    dsci_channel        TEXT NOT NULL,
    dsci_reference      TEXT NOT NULL,
    dsci_version        TEXT NOT NULL,
    effective_at        TIMESTAMPTZ NOT NULL,
    expires_at          TIMESTAMPTZ NULL,
    revoked_at          TIMESTAMPTZ NULL,
    status              TEXT NOT NULL CHECK (status IN ('pending', 'active', 'expired', 'revoked')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID NULL REFERENCES domains(id) ON DELETE SET NULL,

    -- Constraints
    CONSTRAINT check_expiry_after_effective CHECK (expires_at IS NULL OR expires_at > effective_at),
    CONSTRAINT check_revoked_status CHECK (
        (status = 'revoked' AND revoked_at IS NOT NULL) OR
        (status != 'revoked' AND revoked_at IS NULL)
    )
);

-- ============================================================
-- INDEXES
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_contracts_domain_id ON contracts(domain_id);
CREATE INDEX IF NOT EXISTS idx_contracts_subject_domain_id ON contracts(subject_domain_id);
CREATE INDEX IF NOT EXISTS idx_contracts_alias_id ON contracts(alias_id) WHERE alias_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);
CREATE INDEX IF NOT EXISTS idx_contracts_effective_at ON contracts(effective_at);
CREATE INDEX IF NOT EXISTS idx_contracts_created_at ON contracts(created_at DESC);

-- Composite index for common query pattern: find active contracts for a subject
CREATE INDEX IF NOT EXISTS idx_contracts_subject_status ON contracts(subject_domain_id, status) WHERE status = 'active';

-- ============================================================
-- VIEWS
-- ============================================================
-- Active contracts view (convenience for common queries)
CREATE OR REPLACE VIEW active_contracts AS
SELECT
    c.id,
    c.domain_id,
    d.name AS domain_name,
    c.subject_domain_id,
    sd.name AS subject_domain_name,
    c.alias_id,
    a.name AS alias_name,
    c.contract_type,
    c.dsci_channel,
    c.dsci_reference,
    c.dsci_version,
    c.effective_at,
    c.expires_at,
    c.created_at,
    c.created_by
FROM contracts c
JOIN domains d ON d.id = c.domain_id
JOIN domains sd ON sd.id = c.subject_domain_id
LEFT JOIN domains a ON a.id = c.alias_id
WHERE c.status = 'active'
  AND (c.expires_at IS NULL OR c.expires_at > NOW());

-- Contract lineage view (shows contract relationships with domain hierarchy)
CREATE OR REPLACE VIEW contract_lineage AS
SELECT
    c.id,
    c.domain_id,
    d.name AS domain_name,
    d.parent_id AS domain_parent_id,
    c.subject_domain_id,
    sd.name AS subject_domain_name,
    sd.parent_id AS subject_parent_id,
    c.alias_id,
    a.name AS alias_name,
    c.contract_type,
    c.status,
    c.effective_at,
    c.expires_at,
    c.revoked_at,
    c.created_at
FROM contracts c
JOIN domains d ON d.id = c.domain_id
JOIN domains sd ON sd.id = c.subject_domain_id
LEFT JOIN domains a ON a.id = c.alias_id
ORDER BY c.created_at DESC;

-- ============================================================
-- TRIGGERS
-- ============================================================
-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_contracts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_contracts_updated_at
    BEFORE UPDATE ON contracts
    FOR EACH ROW
    EXECUTE FUNCTION update_contracts_updated_at();

-- ============================================================
-- COMMENTS
-- ============================================================
COMMENT ON TABLE contracts IS 'DSCI contracts governing domain participation, membership, and consent relationships (GOV-13).';
COMMENT ON COLUMN contracts.domain_id IS 'Issuing domain (contract owner/offerer).';
COMMENT ON COLUMN contracts.subject_domain_id IS 'Domain subject to the contract (usually a RELATIONSHIP alias domain).';
COMMENT ON COLUMN contracts.alias_id IS 'Optional specific alias domain used when accepting the contract.';
COMMENT ON COLUMN contracts.contract_type IS 'Contract taxonomy: tos, membership, data-processing, subscription, etc.';
COMMENT ON COLUMN contracts.dsci_channel IS 'DSCI channel: web, api, in-person, paper-scan, etc.';
COMMENT ON COLUMN contracts.dsci_reference IS 'Opaque reference to canonical contract artifact (URL, file-id, hash).';
COMMENT ON COLUMN contracts.dsci_version IS 'Contract version string (domain-defined, e.g. v1.0, 2025-11-14).';
COMMENT ON COLUMN contracts.effective_at IS 'Contract becomes effective at this timestamp.';
COMMENT ON COLUMN contracts.expires_at IS 'Optional expiration timestamp (NULL = no expiration).';
COMMENT ON COLUMN contracts.revoked_at IS 'Timestamp when contract was revoked (NULL if not revoked).';
COMMENT ON COLUMN contracts.status IS 'Contract lifecycle status: pending, active, expired, revoked.';
COMMENT ON COLUMN contracts.created_by IS 'Optional: domain/seat that initiated the contract.';

-- ============================================================
-- DOWN MIGRATION
-- ============================================================
-- To rollback this migration:
-- DROP TRIGGER IF EXISTS trigger_contracts_updated_at ON contracts;
-- DROP FUNCTION IF EXISTS update_contracts_updated_at();
-- DROP VIEW IF EXISTS contract_lineage;
-- DROP VIEW IF EXISTS active_contracts;
-- DROP INDEX IF EXISTS idx_contracts_subject_status;
-- DROP INDEX IF EXISTS idx_contracts_created_at;
-- DROP INDEX IF EXISTS idx_contracts_effective_at;
-- DROP INDEX IF EXISTS idx_contracts_status;
-- DROP INDEX IF EXISTS idx_contracts_alias_id;
-- DROP INDEX IF EXISTS idx_contracts_subject_domain_id;
-- DROP INDEX IF EXISTS idx_contracts_domain_id;
-- DROP TABLE IF EXISTS contracts;

-- Verification query (run after migration)
-- SELECT
--     COUNT(*) AS total_contracts,
--     status,
--     contract_type
-- FROM contracts
-- GROUP BY status, contract_type;

-- Bootstrap notification
DO $$
BEGIN
    RAISE NOTICE '✅ GOV-13: Contracts table created successfully';
    RAISE NOTICE '   - Table: contracts with 15 columns';
    RAISE NOTICE '   - Indexes: 7 total (domain_id, subject_domain_id, alias_id, status, effective_at, created_at, subject_status composite)';
    RAISE NOTICE '   - Views: active_contracts, contract_lineage';
    RAISE NOTICE '   - Triggers: updated_at auto-update';
    RAISE NOTICE '   - Ready for DSCI contract wiring';
END $$;
