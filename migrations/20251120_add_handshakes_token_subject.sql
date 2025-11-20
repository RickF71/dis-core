-- +migrate Up
BEGIN;
ALTER TABLE handshakes
  ADD COLUMN IF NOT EXISTS token TEXT;

ALTER TABLE handshakes
  ADD COLUMN IF NOT EXISTS subject TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_handshakes_token_unique ON handshakes(token);
COMMIT;

-- +migrate Down
BEGIN;
DROP INDEX IF EXISTS idx_handshakes_token_unique;
ALTER TABLE handshakes DROP COLUMN IF EXISTS token;
ALTER TABLE handshakes DROP COLUMN IF EXISTS subject;
COMMIT;
