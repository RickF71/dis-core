-- GOV-12: Alias Canon & DSCI Integration
-- Phase: Canonical Alias Taxonomy (AUTO / RELATIONSHIP / MASK)
-- Date: 2025-11-13

-- Create canonical aliases table
CREATE TABLE IF NOT EXISTS aliases (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_domain_id     UUID NOT NULL,
    target_domain_id    UUID NOT NULL,
    alias_name          TEXT NOT NULL,
    alias_type          TEXT NOT NULL CHECK (alias_type IN ('AUTO', 'RELATIONSHIP', 'MASK')),
    is_corporeal_auto   BOOLEAN NOT NULL DEFAULT FALSE,
    dsci_contract_id    UUID NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    retired_at          TIMESTAMPTZ NULL,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Constraints
    CONSTRAINT check_corporeal_auto_must_be_auto
        CHECK (NOT is_corporeal_auto OR alias_type = 'AUTO'),

    -- Unique active alias per scope
    CONSTRAINT unique_active_alias
        EXCLUDE USING btree (owner_domain_id WITH =, target_domain_id WITH =, alias_name WITH =)
        WHERE (retired_at IS NULL)
);

-- Foreign keys
ALTER TABLE aliases
    ADD CONSTRAINT aliases_owner_domain_id_fkey
    FOREIGN KEY (owner_domain_id) REFERENCES domains(id) ON DELETE CASCADE;

ALTER TABLE aliases
    ADD CONSTRAINT aliases_target_domain_id_fkey
    FOREIGN KEY (target_domain_id) REFERENCES domains(id) ON DELETE CASCADE;

-- Note: dsci_contract_id FK will be added when contracts table exists in future phase
-- ALTER TABLE aliases
--     ADD CONSTRAINT aliases_dsci_contract_id_fkey
--     FOREIGN KEY (dsci_contract_id) REFERENCES contracts(id) ON DELETE SET NULL;

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_aliases_owner_domain ON aliases(owner_domain_id, alias_type);
CREATE INDEX IF NOT EXISTS idx_aliases_target_domain ON aliases(target_domain_id, alias_type);
CREATE INDEX IF NOT EXISTS idx_aliases_dsci_contract ON aliases(dsci_contract_id) WHERE dsci_contract_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_aliases_created_at ON aliases(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_aliases_active ON aliases(owner_domain_id, alias_type) WHERE retired_at IS NULL;

-- Comments
COMMENT ON TABLE aliases IS 'GOV-12: Canonical alias taxonomy supporting AUTO (structural), RELATIONSHIP (volitional), and MASK (ephemeral) alias types';
COMMENT ON COLUMN aliases.id IS 'Unique alias identifier';
COMMENT ON COLUMN aliases.owner_domain_id IS 'Domain that owns this alias (typically corporeal or identity root domain)';
COMMENT ON COLUMN aliases.target_domain_id IS 'Domain this alias is scoped to (for RELATIONSHIP/MASK) or represents (for AUTO)';
COMMENT ON COLUMN aliases.alias_name IS 'Human-readable alias handle or identifier';
COMMENT ON COLUMN aliases.alias_type IS 'Alias type: AUTO (structural, non-volitional), RELATIONSHIP (volitional, domain-scoped), MASK (ephemeral, low authority)';
COMMENT ON COLUMN aliases.is_corporeal_auto IS 'True if this is the AUTO alias created for the corporeal root seat (only one per corporeal domain)';
COMMENT ON COLUMN aliases.dsci_contract_id IS 'Future: Link to DSCI contract for seat instantiation (NULL until contracts table exists)';
COMMENT ON COLUMN aliases.created_at IS 'Timestamp when alias was created';
COMMENT ON COLUMN aliases.retired_at IS 'Timestamp when alias was retired (NULL for active aliases)';
COMMENT ON COLUMN aliases.metadata IS 'Additional alias metadata (e.g., TTL for masks, permissions, display preferences)';

-- Create view for active aliases by type
CREATE OR REPLACE VIEW active_aliases_by_type AS
SELECT
    alias_type,
    COUNT(*) AS count,
    COUNT(CASE WHEN is_corporeal_auto THEN 1 END) AS corporeal_auto_count
FROM aliases
WHERE retired_at IS NULL
GROUP BY alias_type;

COMMENT ON VIEW active_aliases_by_type IS 'Summary of active aliases grouped by type';

-- Create view for alias lineage (aliases with domain info)
CREATE OR REPLACE VIEW alias_lineage AS
SELECT
    a.id AS alias_id,
    a.alias_name,
    a.alias_type,
    a.is_corporeal_auto,
    a.created_at,
    a.retired_at,
    a.owner_domain_id,
    d_owner.name AS owner_domain_name,
    a.target_domain_id,
    d_target.name AS target_domain_name,
    CASE
        WHEN a.retired_at IS NULL THEN 'ACTIVE'
        ELSE 'RETIRED'
    END AS status,
    a.metadata
FROM aliases a
LEFT JOIN domains d_owner ON d_owner.id = a.owner_domain_id
LEFT JOIN domains d_target ON d_target.id = a.target_domain_id;

COMMENT ON VIEW alias_lineage IS 'Alias details with resolved domain names for easy inspection';

-- Log migration completion
DO $$
BEGIN
    RAISE NOTICE '✅ GOV-12: Alias Canon & DSCI Integration migration complete';
    RAISE NOTICE '   Created: aliases table with AUTO/RELATIONSHIP/MASK taxonomy';
    RAISE NOTICE '   Indexes: owner_domain, target_domain, dsci_contract, active aliases';
    RAISE NOTICE '   Views: active_aliases_by_type, alias_lineage';
    RAISE NOTICE '   Note: DSCI contract FK pending contracts table (future phase)';
END $$;
