package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB(path string) (*sql.DB, error) {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	schema := `
CREATE TABLE IF NOT EXISTS identities (
	id TEXT PRIMARY KEY,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`
	_, err = db.Exec(schema)

	if err := EnsureReceiptsSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, err
}
