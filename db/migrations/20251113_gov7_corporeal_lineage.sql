-- GOV-7: Corporeal Lineage & Root Seat Establishment
-- Creates canonical corporeal domain hierarchy for human and synthetic entities

-- Phase 1: Ensure terra.numen.lima exists (should already be present)
-- Reference ID: 4daf928e-e58c-454e-8395-f3dedd103dde

-- Phase 2: Create corporeal and corporeal.auto domains
INSERT INTO domains (id, name, parent_id, payload, created_at, updated_at)
VALUES
  -- terra.numen.lima.corporeal — root of embodied, responsible persons
  (
    'a1111111-1111-1111-1111-111111111111',
    'terra.numen.lima.corporeal',
    '4daf928e-e58c-454e-8395-f3dedd103dde', -- parent: terra.numen.lima
    jsonb_build_object(
      'meta', jsonb_build_object(
        'schema_version', 'v1',
        'description', 'Root domain for embodied, responsible persons with physical (terra) and moral (numen) sovereignty',
        'governance', 'human_sovereign',
        'lineage', 'terra.numen.lima'
      ),
      'authority', jsonb_build_object(
        'can_authorize', true,
        'prime_seat', true,
        'self_governing', true
      ),
      'policy', jsonb_build_object(
        'inheritance', 'strict',
        'delegation', 'explicit_only'
      )
    ),
    now(),
    now()
  ),
  -- terra.numen.lima.corporeal.auto — governed synthetic layer
  (
    'a2222222-2222-2222-2222-222222222222',
    'terra.numen.lima.corporeal.auto',
    'a1111111-1111-1111-1111-111111111111', -- parent: corporeal
    jsonb_build_object(
      'meta', jsonb_build_object(
        'schema_version', 'v1',
        'description', 'Synthetic domain for mock users and future AI entities under human governance',
        'governance', 'synthetic_governed',
        'lineage', 'terra.numen.lima.corporeal'
      ),
      'authority', jsonb_build_object(
        'can_authorize', false,
        'requires_proxy', true,
        'inherits_from', 'terra.numen.lima.corporeal'
      ),
      'policy', jsonb_build_object(
        'inheritance', 'strict',
        'self_authorization', false,
        'delegation', 'prohibited'
      )
    ),
    now(),
    now()
  )
ON CONFLICT (id) DO NOTHING;

-- Phase 3: Bind personal domains under corporeal
-- Update user.rick to be under corporeal (real human)
UPDATE domains
SET parent_id = 'a1111111-1111-1111-1111-111111111111',
    payload = payload || jsonb_build_object(
      'meta', COALESCE(payload->'meta', '{}'::jsonb) || jsonb_build_object(
        'governance', 'human_sovereign',
        'parent_domain', 'terra.numen.lima.corporeal'
      )
    ),
    updated_at = now()
WHERE name = 'user.rick'
  AND parent_id IS DISTINCT FROM 'a1111111-1111-1111-1111-111111111111';

-- Phase 4: Bind synthetic domains under corporeal.auto
-- Update user.BOB and user.PAUL to be under corporeal.auto
UPDATE domains
SET parent_id = 'a2222222-2222-2222-2222-222222222222',
    payload = payload || jsonb_build_object(
      'meta', COALESCE(payload->'meta', '{}'::jsonb) || jsonb_build_object(
        'governance', 'synthetic_governed',
        'parent_domain', 'terra.numen.lima.corporeal.auto',
        'synthetic', true
      )
    ),
    updated_at = now()
WHERE name IN ('user.BOB', 'user.PAUL')
  AND parent_id IS DISTINCT FROM 'a2222222-2222-2222-2222-222222222222';

-- Phase 5: Create indexes for lineage queries
CREATE INDEX IF NOT EXISTS idx_domains_parent_id ON domains(parent_id);
CREATE INDEX IF NOT EXISTS idx_domains_name_pattern ON domains(name text_pattern_ops);

-- Phase 6: Create view for corporeal lineage hierarchy
CREATE OR REPLACE VIEW corporeal_lineage AS
WITH RECURSIVE lineage AS (
  -- Base case: corporeal root
  SELECT
    id,
    name,
    parent_id,
    ARRAY[name] as path,
    0 as depth,
    payload->'meta'->>'governance' as governance_type
  FROM domains
  WHERE id = 'a1111111-1111-1111-1111-111111111111'

  UNION ALL

  -- Recursive case: children
  SELECT
    d.id,
    d.name,
    d.parent_id,
    l.path || d.name,
    l.depth + 1,
    d.payload->'meta'->>'governance' as governance_type
  FROM domains d
  INNER JOIN lineage l ON d.parent_id = l.id
)
SELECT
  id,
  name,
  parent_id,
  array_to_string(path, ' → ') as lineage_path,
  depth,
  governance_type,
  CASE
    WHEN governance_type = 'human_sovereign' THEN 'Human'
    WHEN governance_type = 'synthetic_governed' THEN 'Synthetic'
    ELSE 'Unknown'
  END as entity_type
FROM lineage
ORDER BY path;

COMMENT ON VIEW corporeal_lineage IS 'GOV-7: Recursive view showing corporeal domain hierarchy with governance classification';

-- Phase 7: Verification queries (for testing)
DO $$
DECLARE
  corporeal_count int;
  auto_count int;
  human_count int;
  synthetic_count int;
BEGIN
  SELECT COUNT(*) INTO corporeal_count FROM domains WHERE id = 'a1111111-1111-1111-1111-111111111111';
  SELECT COUNT(*) INTO auto_count FROM domains WHERE id = 'a2222222-2222-2222-2222-222222222222';
  SELECT COUNT(*) INTO human_count FROM domains WHERE parent_id = 'a1111111-1111-1111-1111-111111111111';
  SELECT COUNT(*) INTO synthetic_count FROM domains WHERE parent_id = 'a2222222-2222-2222-2222-222222222222';

  RAISE NOTICE '✅ GOV-7 Migration Complete:';
  RAISE NOTICE '   - Corporeal root: % domain(s)', corporeal_count;
  RAISE NOTICE '   - Corporeal.auto: % domain(s)', auto_count;
  RAISE NOTICE '   - Human domains: % domain(s)', human_count;
  RAISE NOTICE '   - Synthetic domains: % domain(s)', synthetic_count;
END $$;
