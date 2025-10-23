package policy

import "database/sql"

func EnsurePoliciesTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS policies (
			name TEXT PRIMARY KEY,
			data JSONB
		)
	`)
	return err
}
