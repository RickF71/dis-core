-- Phase 10J.4: Domain Payload Unification & domain_css Decommission
-- Date: 2025-11-11
-- Purpose: Eliminate redundant domain_css table and flatten domains.data -> domains.payload

-- Step 1: Rename column data to payload
ALTER TABLE domains RENAME COLUMN data TO payload;

-- Step 2: Flatten inner structure (remove extra "data" object while preserving meta and authority)
-- Before: {"data": {"css": {...}, "policy": {...}}, "meta": {...}, "authority": {...}}
-- After: {"css": {...}, "policy": {...}, "meta": {...}, "authority": {...}}
UPDATE domains
SET payload = payload->'data' ||
              jsonb_build_object('meta', COALESCE(payload->'meta', '{}'::jsonb)) ||
              jsonb_build_object('authority', COALESCE(payload->'authority', '{}'::jsonb))
WHERE payload ? 'data';

-- Step 3: Drop the legacy domain_css table (CASCADE to remove dependencies)
DROP TABLE IF EXISTS domain_css CASCADE;
DROP TABLE IF EXISTS domain_css_history CASCADE;

-- Step 4: Add comment explaining the new structure
COMMENT ON COLUMN domains.payload IS 'Unified domain payload with top-level greedy slots: css, policy, receipts, overlay, variables, plus meta and authority';

-- Verification query (run manually after migration):
-- SELECT jsonb_pretty(payload) FROM domains LIMIT 3;
-- Expected keys at top level: css, policy, receipts, overlay, variables, meta, authority
