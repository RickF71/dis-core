-- Migration for authority_decisions table
-- Part of Phase 9B Authority Console integrity activation

CREATE TABLE IF NOT EXISTS authority_decisions (
    id VARCHAR(36) PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    actor VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    policy_id VARCHAR(255) NOT NULL,
    input JSONB NOT NULL,
    result JSONB NOT NULL,
    reason TEXT,
    phase_tag VARCHAR(100),
    replay_hash VARCHAR(64) NOT NULL,

    -- Indexes for common query patterns
    INDEX idx_authority_decisions_domain (domain),
    INDEX idx_authority_decisions_policy (policy_id),
    INDEX idx_authority_decisions_created (created_at),
    INDEX idx_authority_decisions_actor (actor),
    INDEX idx_authority_decisions_phase (phase_tag)
);

-- Add comments for documentation
COMMENT ON TABLE authority_decisions IS 'Stores authority policy decisions for auditability and replay verification';
COMMENT ON COLUMN authority_decisions.id IS 'Unique UUID for the decision';
COMMENT ON COLUMN authority_decisions.actor IS 'Identity/user who triggered the decision';
COMMENT ON COLUMN authority_decisions.domain IS 'Domain context for the decision';
COMMENT ON COLUMN authority_decisions.policy_id IS 'Policy that was evaluated';
COMMENT ON COLUMN authority_decisions.input IS 'Input data provided to policy evaluation';
COMMENT ON COLUMN authority_decisions.result IS 'Result/output of policy evaluation';
COMMENT ON COLUMN authority_decisions.reason IS 'Human-readable explanation of decision';
COMMENT ON COLUMN authority_decisions.phase_tag IS 'Development phase tag for tracking';
COMMENT ON COLUMN authority_decisions.replay_hash IS 'SHA256 hash for replay verification';
