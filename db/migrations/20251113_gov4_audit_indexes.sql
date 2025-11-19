-- GOV-4: Authority decisions table with indexes
-- Tracks all seat transition policy decisions (allowed and denied)

CREATE TABLE IF NOT EXISTS authority_decisions (
    id                  UUID PRIMARY KEY,
    domain_id           TEXT NOT NULL,
    actor_id            TEXT NOT NULL,
    seat                TEXT NOT NULL,           -- terra|numen|lima
    from_state          TEXT NOT NULL,           -- EMPTY|ASSIGNED|OCCUPIED|FROZEN
    to_state            TEXT NOT NULL,           -- EMPTY|ASSIGNED|OCCUPIED|FROZEN
    allow               BOOLEAN NOT NULL,        -- Policy decision
    reason              TEXT NOT NULL,           -- Human-readable explanation
    policy_ref          TEXT NOT NULL,           -- Policy reference (e.g., dis.seat#v1)
    cas_applied         BOOLEAN DEFAULT FALSE,   -- Whether CAS update succeeded
    cas_version_prev    BIGINT DEFAULT 0,        -- Previous version number
    cas_version_new     BIGINT DEFAULT 0,        -- New version number
    request_id          TEXT,                    -- Trace/correlation ID
    extra               JSONB DEFAULT '{}'::jsonb, -- Additional metadata
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- Index for domain-based queries (most recent first)
CREATE INDEX IF NOT EXISTS idx_authority_decisions_domain
    ON authority_decisions(domain_id, created_at DESC);

-- Index for actor-based queries (audit trail)
CREATE INDEX IF NOT EXISTS idx_authority_decisions_actor
    ON authority_decisions(actor_id, created_at DESC);

-- Index for seat-specific queries
CREATE INDEX IF NOT EXISTS idx_authority_decisions_seat
    ON authority_decisions(seat, created_at DESC);

-- Index for policy analysis
CREATE INDEX IF NOT EXISTS idx_authority_decisions_policy
    ON authority_decisions(policy_ref, allow);

-- Index for CAS success rate analysis
CREATE INDEX IF NOT EXISTS idx_authority_decisions_cas
    ON authority_decisions(cas_applied, created_at DESC);

-- Comments for documentation
COMMENT ON TABLE authority_decisions IS 'GOV-4: Authority decision audit log with policy evaluation results';
COMMENT ON COLUMN authority_decisions.cas_applied IS 'True if CAS update succeeded, false if policy denied or CAS conflict';
COMMENT ON COLUMN authority_decisions.policy_ref IS 'REGO policy reference that made the decision';
COMMENT ON COLUMN authority_decisions.request_id IS 'Trace ID for request correlation across services';

-- View for denied transitions (security monitoring)
CREATE OR REPLACE VIEW authority_decisions_denied AS
SELECT
    id,
    domain_id,
    actor_id,
    seat,
    from_state,
    to_state,
    reason,
    policy_ref,
    request_id,
    created_at
FROM authority_decisions
WHERE allow = FALSE
ORDER BY created_at DESC;

COMMENT ON VIEW authority_decisions_denied IS 'Security monitoring: all policy-denied seat transitions';

-- View for CAS conflicts (race condition detection)
CREATE OR REPLACE VIEW authority_decisions_cas_conflicts AS
SELECT
    id,
    domain_id,
    actor_id,
    seat,
    from_state,
    to_state,
    reason,
    cas_version_prev,
    cas_version_new,
    request_id,
    created_at
FROM authority_decisions
WHERE allow = TRUE AND cas_applied = FALSE
ORDER BY created_at DESC;

COMMENT ON VIEW authority_decisions_cas_conflicts IS 'Race condition monitoring: allowed transitions that failed CAS';
