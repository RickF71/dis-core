-- Migration: Add identity_seats table for GOV-1 identity triad (terra/numen/lima)
-- Date: 2025-11-12
-- Phase: GOV-1 Domain Governance Foundation

-- Identity seats table (terra/numen/lima)
CREATE TABLE IF NOT EXISTS identity_seats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_id TEXT NOT NULL,
    seat_type TEXT NOT NULL CHECK (seat_type IN ('terra', 'numen', 'lima')),
    state TEXT NOT NULL DEFAULT 'EMPTY' CHECK (state IN ('EMPTY', 'ASSIGNED', 'OCCUPIED', 'FROZEN')),
    assigned_at TIMESTAMPTZ,
    occupied_at TIMESTAMPTZ,
    frozen_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(identity_id, seat_type)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_identity_seats_identity ON identity_seats(identity_id);
CREATE INDEX IF NOT EXISTS idx_identity_seats_type ON identity_seats(seat_type);
CREATE INDEX IF NOT EXISTS idx_identity_seats_state ON identity_seats(state);

-- View for checking missing triads
CREATE OR REPLACE VIEW identity_triad_status AS
SELECT
    i.id as identity_id,
    MAX(CASE WHEN s.seat_type = 'terra' THEN s.state END) as terra_state,
    MAX(CASE WHEN s.seat_type = 'numen' THEN s.state END) as numen_state,
    MAX(CASE WHEN s.seat_type = 'lima' THEN s.state END) as lima_state
FROM identities i
LEFT JOIN identity_seats s ON i.id = s.identity_id
GROUP BY i.id;

COMMENT ON TABLE identity_seats IS 'Universal identity triad seats (terra/numen/lima) as defined in GOV-1. Terra=existence, Numen=meaning, Lima=consent.';
COMMENT ON COLUMN identity_seats.seat_type IS 'Type of identity seat: terra (existence), numen (meaning), lima (consent)';
COMMENT ON COLUMN identity_seats.state IS 'Seat lifecycle state: EMPTY (no binding), ASSIGNED (bound but inactive), OCCUPIED (active), FROZEN (suspended)';
