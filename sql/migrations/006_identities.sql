-- 006_identities.sql
-- Core identity registry.

CREATE TABLE IF NOT EXISTS identities (
    id         SERIAL PRIMARY KEY,
    dis_uid    TEXT NOT NULL UNIQUE,
    namespace  TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ,
    active     BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_identities_disuid ON identities(dis_uid);
CREATE INDEX IF NOT EXISTS idx_identities_namespace ON identities(namespace);
