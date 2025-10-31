// ...existing code...

// Stub for missing ImportSchema method
// Stub for missing ImportSchema method
// func (l *Ledger) ImportSchema(filename, category, content string) error {
//     // TODO: Implement actual import logic
//     return nil
// }

// Stub for missing InsertReceipt method
// Stub for missing InsertReceipt method
//
//	func (l *Ledger) InsertReceipt(r *Receipt) error {
//	    // TODO: Implement actual DB insert logic
//	    return nil
//	}
//
// AUTOGEN-COPILOT: initial scaffold, verified by RickF71
// Ref: MOAR Phase 2 – Schema Loader
//
// GOAL: Load all .json/.yaml schemas from /schemas directory,
// compute SHA256 for each, verify "version" field, and log results.
// FUNCTIONS: SchemaVerify(), loadSchemaFile(), calcHash()
// RETURN: error if any schema invalid or missing version.
//
// CONTEXT: Supports the DIS MinSet-5 kernel by ensuring all schema files
// are verified before route registration (RegisterAllRoutes()).
package schema

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Receipt struct {
	ReceiptID      string
	By             string
	Action         string
	CreatedAt      time.Time
	Hash           string
	FrozenCoreHash string
	Metadata       map[string]any
}

type ImportReceipt struct {
	ID      string
	Type    string
	Target  string
	Summary string
}

// DefaultSchemaDir defines the canonical relative location for DIS schema files.
// All loaders should build paths from this root to remain portable.
var DefaultSchemaDir = "disyaml/schemas"

// Stub for generateReceiptID
func generateReceiptID() string { return "stub-id" }

// SchemaVerify checks if the schema file has a valid "version" field.
func SchemaVerify(schemaFile string) error {
	data, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to read schema file %s: %w", schemaFile, err)
	}

	var schemaData map[string]any
	if strings.HasSuffix(schemaFile, ".json") {
		if err := json.Unmarshal(data, &schemaData); err != nil {
			return fmt.Errorf("invalid JSON schema %s: %w", schemaFile, err)
		}
	} else if strings.HasSuffix(schemaFile, ".yaml") || strings.HasSuffix(schemaFile, ".yml") {
		if err := yaml.Unmarshal(data, &schemaData); err != nil {
			return fmt.Errorf("invalid YAML schema %s: %w", schemaFile, err)
		}
	} else {
		return fmt.Errorf("unsupported schema format for file %s", schemaFile)
	}

	if _, ok := schemaData["version"]; !ok {
		return fmt.Errorf("missing 'version' field in schema %s", schemaFile)
	}

	return nil
}

// calcHash computes the SHA256 hash of the given content.
func calcHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// UpsertCanonRecord inserts or updates a canon record in the database.
func UpsertCanonRecord(ctx context.Context, db *sql.DB, id, typ, version string, content []byte, sourceFile, hash string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO canon (id, type, version, content, source_file, hash)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			version = EXCLUDED.version,
			content = EXCLUDED.content,
			source_file = EXCLUDED.source_file,
			hash = EXCLUDED.hash,
			imported_at = NOW();`,
		id, typ, version, content, sourceFile, hash)
	if err != nil {
		return fmt.Errorf("store canon: %w", err)
	}
	return nil
}
