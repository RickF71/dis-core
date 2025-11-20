-- +migrate Up
BEGIN;
-- Ensure the domain_seats table has the fields used by K-1 genesis flow
ALTER TABLE domain_seats
  ADD COLUMN IF NOT EXISTS seat_type TEXT NOT NULL DEFAULT 'prime';

ALTER TABLE domain_seats
  ADD COLUMN IF NOT EXISTS member_id TEXT;

ALTER TABLE domain_seats
  ADD COLUMN IF NOT EXISTS appointment_receipt TEXT;

ALTER TABLE domain_seats
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';

CREATE INDEX IF NOT EXISTS idx_domain_seats_type ON domain_seats(seat_type);
CREATE INDEX IF NOT EXISTS idx_domain_seats_member ON domain_seats(member_id);
CREATE INDEX IF NOT EXISTS idx_domain_seats_status ON domain_seats(status);
COMMIT;

-- +migrate Down
BEGIN;
DROP INDEX IF EXISTS idx_domain_seats_status;
DROP INDEX IF EXISTS idx_domain_seats_member;
DROP INDEX IF EXISTS idx_domain_seats_type;
ALTER TABLE domain_seats DROP COLUMN IF EXISTS status;
ALTER TABLE domain_seats DROP COLUMN IF EXISTS appointment_receipt;
ALTER TABLE domain_seats DROP COLUMN IF EXISTS member_id;
ALTER TABLE domain_seats DROP COLUMN IF EXISTS seat_type;
COMMIT;
