-- Migration: Insert `aether` between `null` and `terra`
-- Adds aether as a child of null and reparents terra to aether.
-- Guarded: does nothing if `aether` already exists or if `null`/`terra` are missing.

BEGIN;

-- ensure null and terra exist
DO $$
DECLARE
  null_id TEXT;
  terra_id TEXT;
  aether_exists BOOLEAN;
  aether_id TEXT;
BEGIN
  SELECT id::text INTO null_id FROM domains WHERE name = 'null' OR name = 'domain.null' LIMIT 1;
  IF null_id IS NULL THEN
    RAISE NOTICE 'migration abort: null domain not found';
    RETURN;
  END IF;

  SELECT id::text INTO terra_id FROM domains WHERE name = 'terra' OR name = 'domain.terra' LIMIT 1;
  IF terra_id IS NULL THEN
    RAISE NOTICE 'migration abort: terra domain not found';
    RETURN;
  END IF;

  SELECT EXISTS(SELECT 1 FROM domains WHERE name = 'aether' OR name = 'domain.aether') INTO aether_exists;
  IF aether_exists THEN
    RAISE NOTICE 'aether already exists; skipping';
    RETURN;
  END IF;

  -- insert aether as child of null
  INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at, updated_at)
  VALUES (gen_random_uuid(), 'aether', null_id::uuid, 'aether', '{}'::jsonb, now(), now())
  RETURNING id::text INTO aether_id;

  -- reparent terra to aether
  UPDATE domains SET parent_id = aether_id::uuid
  WHERE name = 'terra' OR name = 'domain.terra';

  RAISE NOTICE 'inserted aether (%), reparented terra (%)', aether_id, terra_id;
END$$;

COMMIT;
