-- +migrate Up
BEGIN;
-- Ensure receipts table has payload and timestamps used by authority receipts
ALTER TABLE receipts
  ADD COLUMN IF NOT EXISTS payload JSONB DEFAULT '{}'::jsonb;

ALTER TABLE receipts
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT now();

ALTER TABLE receipts
  ADD COLUMN IF NOT EXISTS actor TEXT;

ALTER TABLE receipts
  ADD COLUMN IF NOT EXISTS target TEXT;

CREATE INDEX IF NOT EXISTS idx_receipts_type ON receipts(type);
CREATE INDEX IF NOT EXISTS idx_receipts_domain ON receipts(domain);
COMMIT;

-- +migrate Down
BEGIN;
DROP INDEX IF EXISTS idx_receipts_domain;
DROP INDEX IF EXISTS idx_receipts_type;
ALTER TABLE receipts DROP COLUMN IF EXISTS target;
ALTER TABLE receipts DROP COLUMN IF EXISTS actor;
ALTER TABLE receipts DROP COLUMN IF EXISTS created_at;
ALTER TABLE receipts DROP COLUMN IF EXISTS payload;
COMMIT;
