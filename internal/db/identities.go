package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Identity represents a row in the identities table.
type Identity struct {
	ID        int64   `json:"id"`
	DISUID    string  `json:"dis_uid"`
	Namespace string  `json:"namespace,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt *string `json:"updated_at,omitempty"`
	Active    bool    `json:"active"`
}

// EnsureIdentitiesSchema creates the identities table if missing.
func EnsureIdentitiesSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS identities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		dis_uid TEXT UNIQUE NOT NULL,
		namespace TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT,
		active INTEGER DEFAULT 1
	);
	CREATE INDEX IF NOT EXISTS idx_identities_disuid ON identities(dis_uid);
	CREATE INDEX IF NOT EXISTS idx_identities_namespace ON identities(namespace);
	`
	_, err := db.Exec(schema)
	if err == nil {
		fmt.Println("✅ Identities table verified or created.")
	}
	return err
}

// InsertIdentity creates a new identity record.
func InsertIdentity(db *sql.DB, uid string, namespace string) (int64, error) {
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	q := `
	INSERT INTO identities (dis_uid, namespace, created_at, active)
	VALUES (?, ?, ?, 1);
	`
	res, err := db.Exec(q, uid, namespace, ts)
	if err != nil {
		return 0, fmt.Errorf("failed to insert identity: %w", err)
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// GetIdentity retrieves a specific identity by DISUID.
func GetIdentity(db *sql.DB, uid string) (*Identity, error) {
	row := db.QueryRow(`
		SELECT id, dis_uid, namespace, created_at, updated_at, active
		FROM identities
		WHERE dis_uid = ?;
	`, uid)

	var ident Identity
	var updated sql.NullString
	err := row.Scan(
		&ident.ID,
		&ident.DISUID,
		&ident.Namespace,
		&ident.CreatedAt,
		&updated,
		&ident.Active,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}
	if updated.Valid {
		ident.UpdatedAt = &updated.String
	}
	return &ident, nil
}

// ListIdentities returns all active identities with optional limit/offset.
func ListIdentities(db *sql.DB, limit, offset int) ([]Identity, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.Query(`
		SELECT id, dis_uid, namespace, created_at, updated_at, active
		FROM identities
		WHERE active = 1
		ORDER BY id DESC
		LIMIT ? OFFSET ?;
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list identities: %w", err)
	}
	defer rows.Close()

	var list []Identity
	for rows.Next() {
		var ident Identity
		var updated sql.NullString
		if err := rows.Scan(
			&ident.ID,
			&ident.DISUID,
			&ident.Namespace,
			&ident.CreatedAt,
			&updated,
			&ident.Active,
		); err != nil {
			return nil, err
		}
		if updated.Valid {
			ident.UpdatedAt = &updated.String
		}
		list = append(list, ident)
	}
	return list, rows.Err()
}

// DeactivateIdentity marks an identity as inactive and updates timestamp.
func DeactivateIdentity(db *sql.DB, uid string) error {
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.Exec(`
		UPDATE identities
		SET active = 0, updated_at = ?
		WHERE dis_uid = ?;
	`, ts, uid)
	if err == nil {
		fmt.Printf("⚠️  Identity deactivated: %s\n", uid)
	}
	return err
}
