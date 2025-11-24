// postgres_setup.go
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Postgres driver
)

// CreateSchema lays down all base DIS-CORE tables for PostgreSQL.
func CreateSchema(db *sql.DB) error {
	var err error
	schema := []string{
		`CREATE TABLE IF NOT EXISTS receipts (
		       id SERIAL PRIMARY KEY,
		       receipt_id TEXT UNIQUE NOT NULL,
		       schema_ref TEXT,
		       content TEXT,
		       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	       );`,

		`CREATE TABLE IF NOT EXISTS revocations (
				id SERIAL PRIMARY KEY,
				revocation_id TEXT UNIQUE NOT NULL,
				revoked_ref TEXT NOT NULL,
				revoked_type TEXT NOT NULL,
				reason TEXT,
				revoked_by TEXT,
				revocation_time TIMESTAMPTZ,
				valid_until TIMESTAMPTZ,
				signature TEXT
			);`,

		// Updated domains table schema (Phase 10J.4: uses 'payload' instead of 'data')
		`CREATE TABLE IF NOT EXISTS domains (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				parent_id TEXT,
				payload JSONB DEFAULT '{}'::jsonb,  -- Phase 10J.4: renamed from 'data'
				created_at TIMESTAMPTZ DEFAULT now(),
				updated_at TIMESTAMPTZ,
				CONSTRAINT fk_domains_parent
					FOREIGN KEY (parent_id) REFERENCES domains(id) ON DELETE SET NULL
		);`,

		`CREATE TABLE IF NOT EXISTS handshakes (
				id SERIAL PRIMARY KEY,
				handshake_id TEXT UNIQUE NOT NULL,
				initiator TEXT,
				responder TEXT,
				scope TEXT,
				consent_proof TEXT,
				result_token TEXT,
				expires_at TIMESTAMPTZ
			);`,

		`CREATE TABLE IF NOT EXISTS identities (
				id SERIAL PRIMARY KEY,
				dis_uid TEXT UNIQUE NOT NULL,
				namespace TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ,
				active BOOLEAN DEFAULT TRUE
			);`,
	}

	for _, stmt := range schema {
		if _, err = db.Exec(stmt); err != nil {
			return fmt.Errorf("schema creation failed: %w", err)
		}
	}

	// --- Import Warnings Table ---
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS import_warnings (
			id SERIAL PRIMARY KEY,
			file_path TEXT,
			type TEXT,
			domain TEXT,
			schema_type TEXT,
			details TEXT,
			resolved BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			resolved_at TIMESTAMPTZ
	);
	`)
	if err != nil {
		return fmt.Errorf("create import_warnings table: %w", err)
	}
	fmt.Println("✅ import_warnings table ready")

	fmt.Println("✅ PostgreSQL schema initialized.")
	return nil
}

// SeedDefaults inserts baseline domains for the DIS network.
func SeedDefaults(db *sql.DB) error {
	// Insert aether between null and terra where possible, and ensure baseline domains exist.
	_, err := db.Exec(`
	INSERT INTO domains (id, name, parent_id, payload, created_at)
	VALUES
		('domain.null', 'domain.null', NULL, '{}'::jsonb, now())
	ON CONFLICT (id) DO NOTHING;
	`)
	if err != nil {
		return err
	}

	// Ensure aether exists as child of null (if not already present)
	_, err = db.Exec(`
	INSERT INTO domains (id, name, parent_id, payload, created_at)
	VALUES (gen_random_uuid(), 'aether', (SELECT id FROM domains WHERE name = 'domain.null' OR name = 'null' LIMIT 1), '{}'::jsonb, now())
	ON CONFLICT (name) DO NOTHING;
	`)
	if err != nil {
		return err
	}

	// Ensure terra exists and is parented under aether when possible
	_, err = db.Exec(`
	INSERT INTO domains (id, name, parent_id, payload, created_at)
	VALUES
		('domain.terra', 'domain.terra', (SELECT id FROM domains WHERE name = 'aether' OR name = 'domain.aether' LIMIT 1), '{}'::jsonb, now()),
		('domain.virtual.usa', 'domain.virtual.usa', NULL, '{}'::jsonb, now())
	ON CONFLICT (id) DO NOTHING;
	`)
	if err == nil {
		fmt.Println("🌱 Seeded baseline domains: domain.null, aether, domain.terra, domain.virtual.usa")
	}
	return err
}

// Setup initializes PostgreSQL schema and seeds baseline data.
func Setup(db *sql.DB) error {
	if err := CreateSchema(db); err != nil {
		return err
	}
	if err := SeedDefaults(db); err != nil {
		return err
	}
	return nil
}
