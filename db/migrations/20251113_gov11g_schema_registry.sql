-- GOV-11G: Schema Registry & Domain Schema Adoption
-- Phase: Schema-Aware Identity Policy Editing
-- Date: 2025-01-14

-- Create schemas table for registry of available identity schemas
CREATE TABLE IF NOT EXISTS schemas (
    id TEXT NOT NULL,
    version TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (id, version)
);

CREATE INDEX IF NOT EXISTS idx_schemas_id ON schemas(id);
CREATE INDEX IF NOT EXISTS idx_schemas_created_at ON schemas(created_at DESC);

COMMENT ON TABLE schemas IS 'Registry of identity policy schemas available for domain adoption (GOV-11G)';
COMMENT ON COLUMN schemas.id IS 'Schema identifier (e.g., "identity.basic.v1")';
COMMENT ON COLUMN schemas.version IS 'Schema version (e.g., "1.0.0")';
COMMENT ON COLUMN schemas.payload IS 'JSON Schema definition including fields, types, constraints';
COMMENT ON COLUMN schemas.created_at IS 'Schema registration timestamp';

-- Create domain_schemas table for tracking schema adoption per domain
CREATE TABLE IF NOT EXISTS domain_schemas (
    domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    schema_id TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    adopted_at TIMESTAMPTZ DEFAULT NOW(),
    adopted_by UUID,
    metadata JSONB DEFAULT '{}'::jsonb,
    PRIMARY KEY (domain_id, schema_id, schema_version),
    FOREIGN KEY (schema_id, schema_version) REFERENCES schemas(id, version)
);

CREATE INDEX IF NOT EXISTS idx_domain_schemas_domain_id ON domain_schemas(domain_id);
CREATE INDEX IF NOT EXISTS idx_domain_schemas_schema ON domain_schemas(schema_id, schema_version);
CREATE INDEX IF NOT EXISTS idx_domain_schemas_adopted_at ON domain_schemas(adopted_at DESC);

COMMENT ON TABLE domain_schemas IS 'Tracks schema adoption by domains - adopted schemas allow policy editing (GOV-11G)';
COMMENT ON COLUMN domain_schemas.domain_id IS 'Domain that adopted the schema';
COMMENT ON COLUMN domain_schemas.schema_id IS 'Adopted schema ID';
COMMENT ON COLUMN domain_schemas.schema_version IS 'Adopted schema version';
COMMENT ON COLUMN domain_schemas.adopted_at IS 'Timestamp when schema was adopted';
COMMENT ON COLUMN domain_schemas.adopted_by IS 'Actor UUID who initiated adoption (optional)';
COMMENT ON COLUMN domain_schemas.metadata IS 'Additional adoption metadata (overrides, options, etc.)';

-- View for schema set resolution (domain + inherited schemas)
CREATE OR REPLACE VIEW domain_schema_sets AS
WITH RECURSIVE parent_chain AS (
    -- Base case: start with each domain
    SELECT
        d.id AS domain_id,
        d.id AS current_domain_id,
        d.parent_id,
        0 AS depth
    FROM domains d

    UNION ALL

    -- Recursive case: walk up parent chain
    SELECT
        pc.domain_id,
        d.id AS current_domain_id,
        d.parent_id,
        pc.depth + 1 AS depth
    FROM parent_chain pc
    JOIN domains d ON d.id = pc.parent_id
    WHERE pc.parent_id IS NOT NULL AND pc.depth < 10  -- prevent infinite loops
)
SELECT
    pc.domain_id,
    ds.schema_id,
    ds.schema_version,
    pc.current_domain_id AS source_domain_id,
    CASE
        WHEN pc.current_domain_id = pc.domain_id THEN 'adopted'
        ELSE 'inherited'
    END AS mode,
    pc.depth AS inheritance_depth,
    ds.adopted_at,
    ds.adopted_by,
    s.payload AS schema_payload
FROM parent_chain pc
JOIN domain_schemas ds ON ds.domain_id = pc.current_domain_id
JOIN schemas s ON s.id = ds.schema_id AND s.version = ds.schema_version
ORDER BY pc.domain_id, pc.depth, ds.schema_id;

COMMENT ON VIEW domain_schema_sets IS 'Resolves effective schema set for each domain (adopted + inherited from parent chain) - GOV-11G';

-- Insert sample identity schemas for testing
INSERT INTO schemas (id, version, payload) VALUES
(
    'identity.basic.v1',
    '1.0.0',
    '{
        "type": "object",
        "title": "Basic Identity Policy Schema",
        "description": "Core identity policy fields for basic identity management",
        "properties": {
            "allow_foreign_identities": {
                "type": "boolean",
                "description": "Allow acceptance of external identity assertions",
                "default": false
            },
            "require_corporeal_verification": {
                "type": "boolean",
                "description": "Require IRL authentication for identity binding",
                "default": true
            },
            "max_alias_count": {
                "type": "integer",
                "description": "Maximum number of aliases per identity",
                "minimum": 1,
                "maximum": 100,
                "default": 10
            },
            "identity_retention_days": {
                "type": "integer",
                "description": "Days to retain inactive identities",
                "minimum": 30,
                "default": 365
            }
        },
        "required": ["allow_foreign_identities", "require_corporeal_verification"]
    }'::jsonb
),
(
    'identity.extended.v1',
    '1.0.0',
    '{
        "type": "object",
        "title": "Extended Identity Policy Schema",
        "description": "Additional identity policy fields for advanced identity management",
        "properties": {
            "allow_anonymous_access": {
                "type": "boolean",
                "description": "Allow anonymous access without identity binding",
                "default": false
            },
            "trusted_domains": {
                "type": "array",
                "description": "List of trusted domain IDs for identity federation",
                "items": {
                    "type": "string",
                    "format": "uuid"
                },
                "default": []
            },
            "identity_proof_methods": {
                "type": "array",
                "description": "Accepted identity proof methods",
                "items": {
                    "type": "string",
                    "enum": ["biometric", "passkey", "mobile_app", "totp", "sms"]
                },
                "default": ["passkey", "biometric"]
            }
        }
    }'::jsonb
)
ON CONFLICT (id, version) DO NOTHING;

-- Log migration completion
DO $$
BEGIN
    RAISE NOTICE 'GOV-11G migration complete: schema registry tables created';
    RAISE NOTICE 'Sample schemas inserted: identity.basic.v1, identity.extended.v1';
END $$;
