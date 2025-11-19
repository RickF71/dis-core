-- GOV-11H: Domain Branching, Consent Continuity & DSCI Canonization
-- Phase: Establish branching as canonical domain realignment mechanism
-- Date: 2025-11-13

-- Add branching metadata to domains table
ALTER TABLE domains
  ADD COLUMN IF NOT EXISTS branch_of UUID REFERENCES domains(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS branch_depth INTEGER DEFAULT 0 NOT NULL;

-- Create index for branch queries
CREATE INDEX IF NOT EXISTS idx_domains_branch_of ON domains(branch_of);
CREATE INDEX IF NOT EXISTS idx_domains_branch_depth ON domains(branch_depth);

-- Add comments
COMMENT ON COLUMN domains.branch_of IS 'The domain from which this domain was branched (GOV-11H). NULL for non-branched domains.';
COMMENT ON COLUMN domains.branch_depth IS 'Branch generation depth: 0 for original domains, increments with each branch (GOV-11H).';

-- Add constraint: branch_of must be different from id (no self-branching)
ALTER TABLE domains
  ADD CONSTRAINT check_no_self_branch
  CHECK (branch_of IS NULL OR branch_of != id);

-- Create view for branch lineage traversal
CREATE OR REPLACE VIEW domain_branch_lineage AS
WITH RECURSIVE lineage AS (
  -- Base case: domains that are branches
  SELECT
    d.id AS branch_domain_id,
    d.name AS branch_domain_name,
    d.branch_of AS original_domain_id,
    d_orig.name AS original_domain_name,
    d.branch_depth,
    d.parent_id AS branch_parent_id,
    d_parent.name AS branch_parent_name,
    d.created_at AS branched_at,
    1 AS lineage_depth
  FROM domains d
  LEFT JOIN domains d_orig ON d_orig.id = d.branch_of
  LEFT JOIN domains d_parent ON d_parent.id = d.parent_id
  WHERE d.branch_of IS NOT NULL

  UNION ALL

  -- Recursive case: follow branch_of chain
  SELECT
    l.branch_domain_id,
    l.branch_domain_name,
    d.branch_of AS original_domain_id,
    d_orig.name AS original_domain_name,
    l.branch_depth,
    l.branch_parent_id,
    l.branch_parent_name,
    l.branched_at,
    l.lineage_depth + 1
  FROM lineage l
  JOIN domains d ON d.id = l.original_domain_id
  LEFT JOIN domains d_orig ON d_orig.id = d.branch_of
  WHERE d.branch_of IS NOT NULL AND l.lineage_depth < 10
)
SELECT * FROM lineage;

COMMENT ON VIEW domain_branch_lineage IS 'Recursive view showing complete branch lineage for all branched domains (GOV-11H).';

-- Add contract.branch_inheritance field tracking
-- (Note: contracts table may not exist yet; this is a placeholder comment)
-- ALTER TABLE contracts ADD COLUMN IF NOT EXISTS branch_inheritance BOOLEAN DEFAULT TRUE;
-- COMMENT ON COLUMN contracts.branch_inheritance IS 'Whether seats can be instantiated in branches via DSCI (GOV-11H).';

-- Log migration completion
DO $$
BEGIN
    RAISE NOTICE 'GOV-11H migration complete: Domain branching metadata added';
    RAISE NOTICE 'Fields: branch_of, branch_depth';
    RAISE NOTICE 'View: domain_branch_lineage';
    RAISE NOTICE 'Note: parent_id is now IMMUTABLE after domain creation';
END $$;
