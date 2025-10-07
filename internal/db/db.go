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
	if err != nil { return nil, err }

	schema := `
CREATE TABLE IF NOT EXISTS identities (
	id TEXT PRIMARY KEY,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS receipts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	identity_id TEXT,
	action TEXT,
	by_domain TEXT,
	scope TEXT,
	nonce TEXT,
	timestamp TEXT,
	policy_checksum TEXT,
	signature TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`
	_, err = db.Exec(schema)
	return db, err
}
