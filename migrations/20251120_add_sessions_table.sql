-- UP: add sessions table
CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id UUID NOT NULL,
  domain_id UUID NOT NULL,
  seat_id UUID NOT NULL,
  token TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_actor ON sessions(actor_id);
CREATE INDEX IF NOT EXISTS idx_sessions_domain ON sessions(domain_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);

-- Index to help revoked/session scans
CREATE INDEX IF NOT EXISTS idx_sessions_revoked_at ON sessions(revoked_at);

-- DOWN: drop sessions table
-- To rollback: drop indexes and table
-- DROP INDEX IF EXISTS idx_sessions_token;
-- DROP INDEX IF EXISTS idx_sessions_domain;
-- DROP INDEX IF EXISTS idx_sessions_actor;
-- DROP TABLE IF EXISTS sessions;
