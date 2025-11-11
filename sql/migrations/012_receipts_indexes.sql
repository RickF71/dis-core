-- 012_receipts_indexes.sql
-- Helpful indexes on receipts for lineage lookups.

CREATE INDEX IF NOT EXISTS idx_receipts_actor ON receipts(actor);
CREATE INDEX IF NOT EXISTS idx_receipts_domain ON receipts(domain);
CREATE INDEX IF NOT EXISTS idx_receipts_created_at ON receipts(created_at);
