-- Add default identity bindings for dev users
-- This creates bindings for the dev user list (testuser, rick, alice, bob)
-- Each is bound to terra.numen.lima.corporeal as their home domain

-- Note: The 'key' field is not currently used but required by schema.
-- In production, this would contain cryptographic identity proof.

INSERT INTO identity_bindings (uid, domain, key)
VALUES
  ('testuser', 'terra.numen.lima.corporeal', 'dev-key-testuser'),
  ('rick', 'terra.numen.lima.corporeal', 'dev-key-rick'),
  ('alice', 'terra.numen.lima.corporeal', 'dev-key-alice'),
  ('bob', 'terra.numen.lima.corporeal', 'dev-key-bob')
ON CONFLICT (uid, domain) DO NOTHING;

-- Log binding creation
DO $$
BEGIN
  RAISE NOTICE '[identity_bindings] Created dev user bindings for testuser, rick, alice, bob → terra.numen.lima.corporeal';
END
$$;
