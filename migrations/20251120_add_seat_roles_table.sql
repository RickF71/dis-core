-- UP
CREATE TABLE IF NOT EXISTS seat_roles (
  seat_id UUID NOT NULL,
  role TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (seat_id, role)
);
CREATE INDEX IF NOT EXISTS idx_seat_roles_role ON seat_roles(role);

-- DOWN
DROP TABLE IF EXISTS seat_roles;
DROP INDEX IF EXISTS idx_seat_roles_role;
