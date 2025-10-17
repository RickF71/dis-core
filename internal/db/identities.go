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
		id SERIAL PRIMARY KEY,
		dis_uid TEXT UNIQUE NOT NULL,
		namespace TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
		updated_at TIMESTAMPTZ,
		active BOOLEAN DEFAULT TRUE
	);
	CREATE INDEX IF NOT EXISTS idx_identities_disuid ON identities(dis_uid);
	CREATE INDEX IF NOT EXISTS idx_identities_namespace ON identities(namespace);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to ensure identities table: %w", err)
	}
	fmt.Println("✅ identities table verified or created (Postgres).")
	return nil
}

// InsertIdentity creates a new identity record.
func InsertIdentity(db *sql.DB, uid string, namespace string) (int64, error) {
	q := `
	INSERT INTO identities (dis_uid, namespace, created_at, active)
	VALUES ($1, $2, NOW(), TRUE)
	RETURNING id;
	`
	var id int64
	err := db.QueryRow(q, uid, namespace).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert identity: %w", err)
	}
	fmt.Printf("🆔 new identity created: %s\n", uid)
	return id, nil
}

// UpsertIdentity inserts or updates an identity based on dis_uid.
// If the identity exists, updates namespace or reactivates it.
func UpsertIdentity(db *sql.DB, uid, namespace string, active bool) (int64, error) {
	q := `
	INSERT INTO identities (dis_uid, namespace, created_at, active)
	VALUES ($1, $2, NOW(), $3)
	ON CONFLICT (dis_uid)
	DO UPDATE SET
		namespace = EXCLUDED.namespace,
		active = EXCLUDED.active,
		updated_at = NOW()
	RETURNING id;
	`
	var id int64
	err := db.QueryRow(q, uid, namespace, active).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert identity: %w", err)
	}
	fmt.Printf("♻️  identity upserted: %s\n", uid)
	return id, nil
}

// GetIdentity retrieves a specific identity by DISUID.
func GetIdentity(db *sql.DB, uid string) (*Identity, error) {
	row := db.QueryRow(`
		SELECT id, dis_uid, namespace, created_at, updated_at, active
		FROM identities
		WHERE dis_uid = $1;
	`, uid)

	var ident Identity
	var updated sql.NullTime
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
		ident.UpdatedAt = &updated.Time
	}
	return &ident, nil
}

// FindByNamespace retrieves identities by namespace (partial match allowed).
func FindByNamespace(db *sql.DB, ns string) ([]Identity, error) {
	rows, err := db.Query(`
		SELECT id, dis_uid, namespace, created_at, updated_at, active
		FROM identities
		WHERE namespace ILIKE $1;
	`, "%"+ns+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to find by namespace: %w", err)
	}
	defer rows.Close()

	var results []Identity
	for rows.Next() {
		var ident Identity
		var updated sql.NullTime
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
			ident.UpdatedAt = &updated.Time
		}
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
		WHERE active = TRUE
		ORDER BY id DESC
		LIMIT $1 OFFSET $2;
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list identities: %w", err)
	}
	defer rows.Close()

	var list []Identity
	for rows.Next() {
		var ident Identity
		var updated sql.NullTime
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
			ident.UpdatedAt = &updated.Time
		}
		list = append(list, ident)
	}
	return list, rows.Err()
}

// DeactivateIdentity marks an identity as inactive.
func DeactivateIdentity(db *sql.DB, uid string) error {
	_, err := db.Exec(`
		UPDATE identities
		SET active = FALSE, updated_at = NOW()
		WHERE dis_uid = $1;
	`, uid)
	if err == nil {
		fmt.Printf("⚠️  identity deactivated: %s\n", uid)
	}
	return err
}

// ReactivateIdentity re-enables a previously inactive identity.
func ReactivateIdentity(db *sql.DB, uid string) error {
	_, err := db.Exec(`
		UPDATE identities
		SET active = TRUE, updated_at = NOW()
		WHERE dis_uid = $1;
	`, uid)
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
