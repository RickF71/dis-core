package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Identity represents a record from the identities table.
type Identity struct {
	ID        int64      `json:"id"`
	DISUID    string     `json:"dis_uid"`
	Namespace string     `json:"namespace,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Active    bool       `json:"active"`
}

// EnsureIdentitiesSchema creates the identities table if it doesn't exist.
func EnsureIdentitiesSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS identities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		dis_uid TEXT UNIQUE NOT NULL,
		namespace TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT,
		active INTEGER DEFAULT 1 CHECK(active IN (0,1))
	);
	CREATE INDEX IF NOT EXISTS idx_identities_disuid ON identities(dis_uid);
	CREATE INDEX IF NOT EXISTS idx_identities_namespace ON identities(namespace);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to ensure identities table: %w", err)
	}
	fmt.Println("✅ identities table verified or created.")
	return nil
}

// InsertIdentity creates a new identity record.
func InsertIdentity(db *sql.DB, uid string, namespace string) (int64, error) {
	ts := NowRFC3339Nano()
	q := `
	INSERT INTO identities (dis_uid, namespace, created_at, active)
	VALUES (?, ?, ?, 1);
	`
	res, err := db.Exec(q, uid, namespace, ts)
	if err != nil {
		return 0, fmt.Errorf("failed to insert identity: %w", err)
	}
	id, _ := res.LastInsertId()
	fmt.Printf("🆔 new identity created: %s\n", uid)
	return id, nil
}

// UpsertIdentity inserts or updates an identity based on dis_uid.
// If the identity exists, updates namespace or reactivates it.
func UpsertIdentity(db *sql.DB, uid, namespace string, active bool) (int64, error) {
	existing, _ := GetIdentity(db, uid)
	ts := NowRFC3339Nano()

	if existing != nil {
		_, err := db.Exec(`
			UPDATE identities
			SET namespace = ?, active = ?, updated_at = ?
			WHERE dis_uid = ?;
		`, namespace, boolToInt(active), ts, uid)
		if err != nil {
			return existing.ID, fmt.Errorf("failed to update identity: %w", err)
		}
		fmt.Printf("♻️  identity updated: %s\n", uid)
		return existing.ID, nil
	}
	return InsertIdentity(db, uid, namespace)
}

// GetIdentity retrieves a specific identity by DISUID.
func GetIdentity(db *sql.DB, uid string) (*Identity, error) {
	row := db.QueryRow(`
		SELECT id, dis_uid, namespace, created_at, updated_at, active
		FROM identities
		WHERE dis_uid = ?;
	`, uid)

	var ident Identity
	var created, updated sql.NullString
	var activeInt int
	err := row.Scan(
		&ident.ID,
		&ident.DISUID,
		&ident.Namespace,
		&created,
		&updated,
		&activeInt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	if created.Valid {
		if t, e := time.Parse(time.RFC3339Nano, created.String); e == nil {
			ident.CreatedAt = t
		}
	}
	if updated.Valid {
		if t, e := time.Parse(time.RFC3339Nano, updated.String); e == nil {
			ident.UpdatedAt = &t
		}
	}
	ident.Active = activeInt == 1
	return &ident, nil
}

// FindByNamespace retrieves identities by namespace (partial match allowed).
func FindByNamespace(db *sql.DB, ns string) ([]Identity, error) {
	rows, err := db.Query(`
		SELECT id, dis_uid, namespace, created_at, updated_at, active
		FROM identities
		WHERE namespace LIKE ?;
	`, "%"+ns+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to find by namespace: %w", err)
	}
	defer rows.Close()

	var results []Identity
	for rows.Next() {
		var ident Identity
		var created, updated sql.NullString
		var activeInt int
		if err := rows.Scan(
			&ident.ID,
			&ident.DISUID,
			&ident.Namespace,
			&created,
			&updated,
			&activeInt,
		); err != nil {
			return nil, err
		}
		if created.Valid {
			t, _ := time.Parse(time.RFC3339Nano, created.String)
			ident.CreatedAt = t
		}
		if updated.Valid {
			t, _ := time.Parse(time.RFC3339Nano, updated.String)
			ident.UpdatedAt = &t
		}
		ident.Active = activeInt == 1
		results = append(results, ident)
	}
	return results, rows.Err()
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
		var created, updated sql.NullString
		var activeInt int
		if err := rows.Scan(
			&ident.ID,
			&ident.DISUID,
			&ident.Namespace,
			&created,
			&updated,
			&activeInt,
		); err != nil {
			return nil, err
		}
		if created.Valid {
			t, _ := time.Parse(time.RFC3339Nano, created.String)
			ident.CreatedAt = t
		}
		if updated.Valid {
			t, _ := time.Parse(time.RFC3339Nano, updated.String)
			ident.UpdatedAt = &t
		}
		ident.Active = activeInt == 1
		list = append(list, ident)
	}
	return list, rows.Err()
}

// DeactivateIdentity marks an identity as inactive.
func DeactivateIdentity(db *sql.DB, uid string) error {
	ts := NowRFC3339Nano()
	_, err := db.Exec(`
		UPDATE identities
		SET active = 0, updated_at = ?
		WHERE dis_uid = ?;
	`, ts, uid)
	if err == nil {
		fmt.Printf("⚠️  identity deactivated: %s\n", uid)
	}
	return err
}

// ReactivateIdentity re-enables a previously inactive identity.
func ReactivateIdentity(db *sql.DB, uid string) error {
	ts := NowRFC3339Nano()
	_, err := db.Exec(`
		UPDATE identities
		SET active = 1, updated_at = ?
		WHERE dis_uid = ?;
	`, ts, uid)
	if err == nil {
		fmt.Printf("✅ identity reactivated: %s\n", uid)
	}
	return err
}

// CountIdentities returns the total number of identities in DB.
func CountIdentities() (int64, error) {
	if DefaultDB == nil {
		return 0, fmt.Errorf("db not initialized")
	}

	var n int64
	err := DefaultDB.QueryRow(`SELECT COUNT(1) FROM identities;`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("failed to count identities: %w", err)
	}
	return n, nil
}

// Helper
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
