package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SetupDatabase ensures the SQLite DB file exists and all tables are created.
// func SetupDatabase(path string) (*sql.DB, error) {
// 	firstCreate := false
// 	if _, err := os.Stat(path); os.IsNotExist(err) {
// 		firstCreate = true
// 		fmt.Printf("🧩 No database found at %s — creating a new one...\n", path)
// 	}

// 	db, err := sql.Open("sqlite", path)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open database: %w", err)
// 	}

// 	// --- Initial creation path ---
// 	if firstCreate {
// 		if err := createSchema(db); err != nil {
// 			return nil, fmt.Errorf("failed to initialize schema: %w", err)
// 		}
// 		if err := seedDefaults(db); err != nil {
// 			return nil, fmt.Errorf("failed to seed defaults: %w", err)
// 		}
// 	}

// 	// --- Always ensure critical schemas exist ---
// 	if err := EnsureReceiptsSchema(db); err != nil {
// 		return nil, err
// 	}
// 	if err := EnsureIdentitiesSchema(db); err != nil {
// 		return nil, err
// 	}

// 	fmt.Println("✅ Database ready.")
// 	return db, nil
// }

// createSchema lays down all base DIS-CORE tables.
func createSchema(db *sql.DB) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS receipts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			receipt_id TEXT UNIQUE NOT NULL,
			schema_ref TEXT,
			content TEXT,
			timestamp TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS revocations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			revocation_id TEXT UNIQUE NOT NULL,
			revoked_ref TEXT NOT NULL,
			revoked_type TEXT NOT NULL,
			reason TEXT,
			revoked_by TEXT,
			revocation_time TEXT,
			valid_until TEXT,
			signature TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain_name TEXT UNIQUE NOT NULL,
			schema_ref TEXT,
			metadata TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS handshakes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			handshake_id TEXT UNIQUE NOT NULL,
			initiator TEXT,
			responder TEXT,
			scope TEXT,
			consent_proof TEXT,
			result_token TEXT,
			expires_at TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS identities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dis_uid TEXT UNIQUE NOT NULL,
			namespace TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT,
			active INTEGER DEFAULT 1
		);`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("schema creation failed: %w", err)
		}
	}

	fmt.Println("✅ SQLite schema initialized.")
	return nil
}

// seedDefaults inserts baseline domains for the DIS network.
func seedDefaults(db *sql.DB) error {
	_, err := db.Exec(`
        INSERT OR IGNORE INTO domains (domain_name, schema_ref, metadata)
        VALUES 
            ('domain.null', 'dis-core.v1', '{}'),
            ('domain.terra', 'dis-core.v1', '{}'),
            ('domain.virtual.usa', 'virtual_usa.credential.v0', '{}');
    `)
	if err == nil {
		fmt.Println("🌱 Seeded baseline domains: domain.null, domain.terra, domain.virtual.usa")
	}
	return err
}
