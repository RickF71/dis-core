package ledger

import (
	"context"
	"crypto/sha256"
	"dis-core/internal/schema"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Ledger provides the core persistence layer for DIS-Core.
// It manages schema creation, config, canon, and event logging.
type Ledger struct {
	DB  *pgx.Conn
	reg *schema.Registry
}

// Open initializes the ledger. It can accept either an existing DB handle
// or a DSN string. If db is nil, it opens a new connection using the DSN.
// Optionally, a schema registry can be attached for validation and linkage.
func Open(ctx context.Context, dsn string, db *pgx.Conn, reg *schema.Registry) (*Ledger, error) {
	var conn *pgx.Conn
	var err error

	if db != nil {
		conn = db
	} else {
		conn, err = pgx.Connect(ctx, dsn)
		if err != nil {
			return nil, fmt.Errorf("open ledger db: %w", err)
		}
	}

	// Ensure schema tables exist
	schemaStatements := []string{
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

	for _, stmt := range schemaStatements {
		if _, err := conn.Exec(context.Background(), stmt); err != nil {
			return nil, fmt.Errorf("create tables: %w", err)
		}
	}

	// Build the ledger instance and attach registry
	l := &Ledger{
		DB:  conn,
		reg: reg, // <— attach live schema registry
	}

	return l, nil
}

// Close cleanly shuts down the ledger database connection.
func (l *Ledger) Close(ctx context.Context) error {
	return l.DB.Close(ctx)
}

// Record inserts a generic event receipt into the ledger.
func (l *Ledger) Record(ctx context.Context, eventType string, payload map[string]any) error {
	j, _ := json.Marshal(payload)
	_, err := l.DB.Exec(ctx, `
		INSERT INTO receipts (type, payload)
		VALUES ($1, $2);`, eventType, string(j))
	if err != nil {
		return fmt.Errorf("record event: %w", err)
	}
	return nil
}

// StoreCanon inserts or updates a canonical record from a parsed YAML.
func (l *Ledger) StoreCanon(ctx context.Context, rec any) error {
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

	_, err := l.DB.Exec(ctx, stmt,
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
func (l *Ledger) SetConfig(ctx context.Context, key, value string) error {
	_, err := l.DB.Exec(ctx, `
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
func (l *Ledger) GetConfig(ctx context.Context, key string) (string, error) {
	var value string
	err := l.DB.QueryRow(ctx, `SELECT value FROM config WHERE key = $1;`, key).Scan(&value)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return "", nil
		}
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

// Registry returns the schema registry associated with this ledger.
func (l *Ledger) Registry() *schema.Registry {
	if l == nil {
		return nil
	}
	return l.reg
}
