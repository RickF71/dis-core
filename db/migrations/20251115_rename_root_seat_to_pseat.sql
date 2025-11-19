-- DIS Prime Seat Migration
-- Rename "root seat" sovereignty concept to "Prime Seat" (pseat)
-- Date: 2025-11-15
-- Author: DIS System Migration

-- IMPORTANT: This migration affects ONLY sovereignty seat logic.
-- It does NOT touch structural roots like domain.null root domain.

BEGIN;

-- Step 1: Rename seat_type 'root' to 'prime' in domain_seats table
UPDATE domain_seats
SET seat_type = 'prime'
WHERE seat_type = 'root';

-- Step 2: Update seat_type variations (root_corporeal, root_actor, etc.)
UPDATE domain_seats
SET seat_type = 'prime'
WHERE seat_type LIKE 'root_%';

-- Step 3: Update any existing receipts that reference seat="root" to seat="prime"
-- (If receipts table has a seat field in metadata JSONB)
UPDATE receipts
SET metadata = jsonb_set(metadata, '{seat}', '"prime"'::jsonb)
WHERE metadata->>'seat' = 'root';

-- Step 4: Verify migration
DO $$
DECLARE
    prime_seat_count INTEGER;
    old_root_count INTEGER;
    root_variant_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO prime_seat_count FROM domain_seats WHERE seat_type = 'prime';
    SELECT COUNT(*) INTO old_root_count FROM domain_seats WHERE seat_type = 'root';
    SELECT COUNT(*) INTO root_variant_count FROM domain_seats WHERE seat_type LIKE 'root_%';

    RAISE NOTICE 'Migration complete: % Prime Seats found', prime_seat_count;
    RAISE NOTICE 'Old root seats remaining: %', old_root_count;
    RAISE NOTICE 'Old root_* variants remaining: %', root_variant_count;

    IF old_root_count > 0 OR root_variant_count > 0 THEN
        RAISE WARNING 'Found seats still with root/root_* types. Migration may be incomplete.';
    END IF;
END $$;

COMMIT;

-- Post-migration notes:
-- - All seat_type='root' changed to seat_type='prime'
-- - Receipt metadata updated where seat='root' → seat='prime'
-- - Go code must be updated to use 'prime' instead of 'root' for seat types
-- - Rego policies must check seat_type='prime' instead of seat_type='root'
-- - API routes must change /seat/root to /seat/prime
-- - Frontend must display "Prime Seat" instead of "Root Seat"
