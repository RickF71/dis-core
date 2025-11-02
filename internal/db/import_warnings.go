package db

import (
	"database/sql"
	"fmt"
)

// EnsureImportWarningsTable ensures the import_warnings table exists
func EnsureImportWarningsTable(db *sql.DB) error {
	_, err := db.Exec(`
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
	return nil
}
