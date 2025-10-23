package domain

import "database/sql"

func EnsureDomainsTable(db *sql.DB) error {
	_, err := db.Exec(`
	       CREATE EXTENSION IF NOT EXISTS pgcrypto;

	       CREATE TABLE IF NOT EXISTS domains (
		       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		       parent_id UUID,
		       name TEXT NOT NULL UNIQUE,
		       data JSONB DEFAULT '{}'::jsonb,
		       is_notech BOOLEAN NOT NULL DEFAULT FALSE,
		       requires_inside_domain BOOLEAN NOT NULL DEFAULT TRUE,
		       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		       FOREIGN KEY (parent_id) REFERENCES domains(id)
	       );
       `)
	return err
}
