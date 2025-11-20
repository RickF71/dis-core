-- +migrate Up
BEGIN;
ALTER TABLE identities
  ADD COLUMN IF NOT EXISTS presentation_name TEXT;

ALTER TABLE identities
  ADD COLUMN IF NOT EXISTS identity_type TEXT;

CREATE INDEX IF NOT EXISTS idx_identities_presentation_name ON identities(presentation_name);
COMMIT;

-- +migrate Down
BEGIN;
DROP INDEX IF EXISTS idx_identities_presentation_name;
ALTER TABLE identities DROP COLUMN IF EXISTS presentation_name;
ALTER TABLE identities DROP COLUMN IF EXISTS identity_type;
COMMIT;
