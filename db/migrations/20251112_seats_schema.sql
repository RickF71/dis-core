-- Phase S1: Seats Schema Migration
-- Creates domain_seats table for root and member seat management

CREATE TABLE IF NOT EXISTS domain_seats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    parent_seat_id UUID REFERENCES domain_seats(id) ON DELETE SET NULL,
    seat_type TEXT NOT NULL DEFAULT 'root',
    member_id UUID REFERENCES identities(id) ON DELETE SET NULL,
    appointed_by UUID REFERENCES domain_seats(id) ON DELETE SET NULL,
    appointment_receipt TEXT,
    rego_ref TEXT,
    rego_text TEXT,
    policy_version TEXT,
    scope TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT now()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_domain_seats_domain ON domain_seats(domain_id);
CREATE INDEX IF NOT EXISTS idx_domain_seats_type ON domain_seats(seat_type);
CREATE INDEX IF NOT EXISTS idx_domain_seats_status ON domain_seats(status);
CREATE INDEX IF NOT EXISTS idx_domain_seats_member ON domain_seats(member_id);
CREATE INDEX IF NOT EXISTS idx_domain_seats_parent ON domain_seats(parent_seat_id);

-- Comments for documentation
COMMENT ON TABLE domain_seats IS 'Phase S1: Stores root and member seats for domain authority delegation';
COMMENT ON COLUMN domain_seats.seat_type IS 'root | member';
COMMENT ON COLUMN domain_seats.status IS 'active | frozen | detached';
COMMENT ON COLUMN domain_seats.rego_text IS 'Per-seat REGO policy (package dis.seat.<seat_id>)';
