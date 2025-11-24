ALTER TABLE receipts
  ADD COLUMN IF NOT EXISTS action_panel jsonb,
  ADD COLUMN IF NOT EXISTS policy_panel jsonb,
  ADD COLUMN IF NOT EXISTS identity_panel jsonb,
  ADD COLUMN IF NOT EXISTS dimension_panel jsonb,
  ADD COLUMN IF NOT EXISTS lineage_panel jsonb,
  ADD COLUMN IF NOT EXISTS domain_panel jsonb;

-- Ensure indexes for origin/panels if desired (optional)
CREATE INDEX IF NOT EXISTS idx_receipts_origin_domain_id ON receipts(origin_domain_id);
CREATE INDEX IF NOT EXISTS idx_receipts_action_panel ON receipts USING gin (action_panel jsonb_path_ops);
