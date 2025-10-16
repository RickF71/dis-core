package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Handshake represents a session or token exchange.
type Handshake struct {
	ID        int64
	Token     string
	Subject   string
	Initiator string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// EnsureHandshakesSchema creates the handshakes table if needed.
func EnsureHandshakesSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS handshakes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		token TEXT UNIQUE NOT NULL,
		subject TEXT NOT NULL,
		initiator TEXT,
		expires_at TEXT,
		revoked_at TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_handshakes_token ON handshakes(token);
	`
	_, err := db.Exec(schema)
	return err
}

// ListExpiredActiveHandshakes finds expired and not-yet-revoked tokens.
func ListExpiredActiveHandshakes(now time.Time) ([]Handshake, error) {
	rows, err := DefaultDB.Query(`
		SELECT id, token, subject, initiator, expires_at, revoked_at
		FROM handshakes
		WHERE revoked_at IS NULL AND expires_at IS NOT NULL AND expires_at < ?;
	`, now.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("query expired handshakes: %w", err)
	}
	defer rows.Close()

	var list []Handshake
	for rows.Next() {
		var hs Handshake
		var expires, revoked sql.NullString
		if err := rows.Scan(&hs.ID, &hs.Token, &hs.Subject, &hs.Initiator, &expires, &revoked); err != nil {
			return nil, err
		}
		if expires.Valid {
			hs.ExpiresAt, _ = time.Parse(time.RFC3339, expires.String)
		}
		if revoked.Valid {
			t, _ := time.Parse(time.RFC3339, revoked.String)
			hs.RevokedAt = &t
		}
		list = append(list, hs)
	}
	return list, rows.Err()
}

// MarkHandshakeRevoked marks a handshake as revoked.
func MarkHandshakeRevoked(id int64, when time.Time, reason string) error {
	_, err := DefaultDB.Exec(`
		UPDATE handshakes
		SET revoked_at = ?
		WHERE id = ?;
	`, when.Format(time.RFC3339), id)
	if err == nil {
		fmt.Printf("🔒 Handshake %d revoked (%s)\n", id, reason)
	}
	return err
}

// GetHandshakeByToken retrieves a single handshake record by token.
func GetHandshakeByToken(token string) (Handshake, error) {
	var hs Handshake
	var expires, revoked sql.NullString
	row := DefaultDB.QueryRow(`
		SELECT id, token, subject, initiator, expires_at, revoked_at
		FROM handshakes
		WHERE token = ?;
	`, token)
	err := row.Scan(&hs.ID, &hs.Token, &hs.Subject, &hs.Initiator, &expires, &revoked)
	if err != nil {
		return hs, err
	}
	if expires.Valid {
		hs.ExpiresAt, _ = time.Parse(time.RFC3339, expires.String)
	}
	if revoked.Valid {
		t, _ := time.Parse(time.RFC3339, revoked.String)
		hs.RevokedAt = &t
	}
	return hs, nil
}

// CountHandshakes returns total handshake records.
func CountHandshakes() (int64, error) {
	var n int64
	err := DefaultDB.QueryRow(`SELECT COUNT(1) FROM handshakes;`).Scan(&n)
	return n, err
}

// CountRevocations returns number of revoked handshakes.
func CountRevocations() (int64, error) {
	var n int64
	err := DefaultDB.QueryRow(`SELECT COUNT(1) FROM handshakes WHERE revoked_at IS NOT NULL;`).Scan(&n)
	return n, err
}
