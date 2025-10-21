package ledger

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Ledger provides the core persistence layer for DIS-Core.
// It manages schema creation, config, canon, and event logging.
type Ledger struct {
	DB *sql.DB
}

// Open initializes the ledger. It can accept either an existing DB handle
// or a DSN string. If db is nil, it opens a new connection using the DSN.
// This prevents multiple func Open() collisions across the package.
func Open(dsn string, db *sql.DB) (*Ledger, error) {
	var conn *sql.DB
	var err error

	if db != nil {
		conn = db
	} else {
		conn, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("open ledger db: %w", err)
		}
	}

	// Ensure schema exists
	schema := []string{
		`CREATE TABLE IF NOT EXISTS receipts (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			type TEXT NOT NULL,
			actor TEXT,
			target TEXT,
			domain TEXT,
			payload JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS canon (
			id TEXT PRIMARY KEY,
			type TEXT,
			version TEXT,
			content JSONB,
			source_file TEXT,
			hash TEXT,
			imported_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT,
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,
	}

	for _, stmt := range schema {
		if _, err := conn.Exec(stmt); err != nil {
			return nil, fmt.Errorf("create tables: %w", err)
		}
	}

	fmt.Println("✅ Ledger ready.")
	return &Ledger{DB: conn}, nil
}

// Close cleanly shuts down the ledger database connection.
func (l *Ledger) Close() error {
	return l.DB.Close()
}

// Record inserts a generic event receipt into the ledger.
func (l *Ledger) Record(eventType string, payload map[string]any) error {
	j, _ := json.Marshal(payload)
	_, err := l.DB.Exec(`
		INSERT INTO receipts (type, payload)
		VALUES ($1, $2);`, eventType, string(j))
	if err != nil {
		return fmt.Errorf("record event: %w", err)
	}
	return nil
}

// StoreCanon inserts or updates a canonical record from a parsed YAML.
func (l *Ledger) StoreCanon(rec any) error {
	b, _ := json.Marshal(rec)

	var r struct {
		ID      string         `json:"id"`
		Type    string         `json:"type"`
		Version string         `json:"version"`
		Content map[string]any `json:"content"`
		Meta    map[string]any `json:"meta"`
		Hash    string         `json:"hash"`
	}
	_ = json.Unmarshal(b, &r)

	stmt := `
		INSERT INTO canon (id, type, version, content, source_file, hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			version = EXCLUDED.version,
			content = EXCLUDED.content,
			source_file = EXCLUDED.source_file,
			hash = EXCLUDED.hash,
			imported_at = NOW();`

	_, err := l.DB.Exec(stmt,
		r.ID,
		r.Type,
		r.Version,
		string(b),
		getMetaString(r.Meta, "source_file"),
		r.Hash,
	)
	if err != nil {
		return fmt.Errorf("store canon: %w", err)
	}

	return nil
}

// SetConfig updates or inserts a configuration key/value pair.
func (l *Ledger) SetConfig(key, value string) error {
	_, err := l.DB.Exec(`
		INSERT INTO config (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			updated_at = NOW();`,
		key, value)
	if err != nil {
		return fmt.Errorf("set config: %w", err)
	}
	return nil
}

// GetConfig retrieves a config value.
func (l *Ledger) GetConfig(key string) (string, error) {
	var value string
	err := l.DB.QueryRow(`SELECT value FROM config WHERE key = $1;`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get config: %w", err)
	}
	return value, nil
}

// NowRFC3339Nano returns the current UTC timestamp in RFC3339Nano format.
func NowRFC3339Nano() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// HashString returns a SHA256 hex digest.
func HashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Helper: safely fetch string from meta map
func getMetaString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
