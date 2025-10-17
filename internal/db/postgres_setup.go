// postgres_setup.go
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Postgres driver
)

// CreateSchema lays down all base DIS-CORE tables for PostgreSQL.
func CreateSchema(db *sql.DB) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS receipts (
			id SERIAL PRIMARY KEY,
			receipt_id TEXT UNIQUE NOT NULL,
			schema_ref TEXT,
			content TEXT,
			timestamp TIMESTAMPTZ DEFAULT NOW()
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

		`CREATE TABLE IF NOT EXISTS domains (
			id SERIAL PRIMARY KEY,
			domain_name TEXT UNIQUE NOT NULL,
			schema_ref TEXT,
			metadata JSONB DEFAULT '{}'::jsonb
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
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("schema creation failed: %w", err)
		}
	}

	fmt.Println("✅ PostgreSQL schema initialized.")
	return nil
}

// SeedDefaults inserts baseline domains for the DIS network.
func SeedDefaults(db *sql.DB) error {
	_, err := db.Exec(`
	INSERT INTO domains (domain_name, schema_ref, metadata)
	VALUES 
		('domain.null', 'dis-core.v1', '{}'::jsonb),
		('domain.terra', 'dis-core.v1', '{}'::jsonb),
		('domain.virtual.usa', 'virtual_usa.credential.v0', '{}'::jsonb)
	ON CONFLICT (domain_name) DO NOTHING;
	`)
	if err == nil {
		fmt.Println("🌱 Seeded baseline domains: domain.null, domain.terra, domain.virtual.usa")
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
