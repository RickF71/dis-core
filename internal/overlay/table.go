package overlay

import "database/sql"

func EnsureOverlaysTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS overlays (
			name TEXT PRIMARY KEY,
			data JSONB
		)
	`)
	return err
}
