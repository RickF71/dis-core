-- Migration: Add sessions table for persistent session tokens (MX-K4)
-- Date: 2025-11-20

-- +migrate Up
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID NOT NULL,
    domain_id UUID NOT NULL,
    seat_id UUID NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL
);

-- Create a unique btree index on token for fast lookups and to enforce uniqueness
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_token_btree ON sessions USING btree (token);

-- Create a non-unique btree index on expires_at to speed expiry scans/cleanup
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at_btree ON sessions USING btree (expires_at);

-- Index to help revoke/list by user
CREATE INDEX IF NOT EXISTS idx_sessions_revoked_at_btree ON sessions USING btree (revoked_at);

-- +migrate Down
-- Drop indexes first, then the table
DROP INDEX IF EXISTS idx_sessions_expires_at_btree;
DROP INDEX IF EXISTS idx_sessions_token_btree;
DROP TABLE IF EXISTS sessions;
