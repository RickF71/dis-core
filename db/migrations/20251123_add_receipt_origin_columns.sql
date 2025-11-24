-- Add origin domain columns to receipts for provenance
ALTER TABLE receipts
  ADD COLUMN IF NOT EXISTS origin_domain_id text,
  ADD COLUMN IF NOT EXISTS origin_domain_name text;

-- Optionally populate origin_domain_name from existing domain column if available
-- (best-effort; may be no-op in many installs)
-- UPDATE receipts SET origin_domain_id = domain WHERE origin_domain_id IS NULL AND domain IS NOT NULL;
