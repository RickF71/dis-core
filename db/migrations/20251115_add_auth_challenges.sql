-- Migration: Create auth_challenges table
-- Purpose: Support QR-based challenge/response authentication
-- Date: 2025-11-15

CREATE TABLE IF NOT EXISTS auth_challenges (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    browser_session_id  TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'authenticated', 'expired')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at          TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '10 minutes'),
    redeemed_by         TEXT,
    payload             JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_auth_challenges_browser_session ON auth_challenges(browser_session_id);
CREATE INDEX IF NOT EXISTS idx_auth_challenges_status ON auth_challenges(status);
CREATE INDEX IF NOT EXISTS idx_auth_challenges_expires_at ON auth_challenges(expires_at);

COMMENT ON TABLE auth_challenges IS 'QR-based authentication challenges for None Space auth flow (Phase 9C)';
COMMENT ON COLUMN auth_challenges.browser_session_id IS 'Browser session identifier (cookie or generated session key)';
COMMENT ON COLUMN auth_challenges.status IS 'Challenge status: pending, authenticated, or expired';
COMMENT ON COLUMN auth_challenges.redeemed_by IS 'User ID that completed the challenge (set on authentication)';
COMMENT ON COLUMN auth_challenges.payload IS 'Optional metadata (QR payload, device info, etc)';
