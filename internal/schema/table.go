package schema

import "database/sql"

func EnsureSchemasTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schemas (
			name TEXT PRIMARY KEY,
			data JSONB
		)
	`)
	return err
}
