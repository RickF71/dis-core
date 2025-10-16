package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DefaultDB *sql.DB // ✅ global handle for daemon & status API

// SetupDatabase opens (and if needed creates) the SQLite file and ensures all core tables exist.
func SetupDatabase(path string) (*sql.DB, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Sanity check connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// --- Ensure schemas ---
	if err := EnsureReceiptsSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure receipts: %w", err)
	}
	if err := EnsureIdentitiesSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure identities: %w", err)
	}
	if err := EnsureHandshakesSchema(db); err != nil {
		// optional — handshakes.go will provide this
		fmt.Println("⚠️  Handshakes schema not created:", err)
	}

	DefaultDB = db
	fmt.Println("✅ Database ready:", path)
	return db, nil
}

// CloseDatabase safely closes DefaultDB (optional helper)
func CloseDatabase() {
	if DefaultDB != nil {
		_ = DefaultDB.Close()
		DefaultDB = nil
	}
}
