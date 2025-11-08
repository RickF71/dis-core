package schema

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var authoritySchema []byte

// SchemaInfo describes the metadata stored in DB
type SchemaInfo struct {
	ID       string          `db:"id" json:"id"`
	Name     string          `db:"name" json:"name"`
	Version  string          `db:"version" json:"version"`
	JSON     json.RawMessage `db:"json_schema" json:"json_schema"`
	IsActive bool            `db:"is_active" json:"is_active"`
	Hash     string          `db:"hash" json:"hash"`
}

// LoadOrRegister ensures the canonical schema is registered in DB.
func LoadOrRegister(ctx context.Context, db *sql.DB) error {
	hash := calcHash(string(authoritySchema))

	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM schemas
		WHERE name = 'authority.console' AND version = '0.9.3'
	`).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("schema check failed: %w", err)
	}

	if count == 0 {
		_, err := db.ExecContext(ctx, `
			INSERT INTO schemas (name, version, json_schema, is_active)
			VALUES ($1, $2, $3, true)
		`, "authority.console", "0.9.3", authoritySchema)
		if err != nil {
			return fmt.Errorf("failed to insert schema: %w", err)
		}
		fmt.Println("[schema] authority.console v0.9.3 registered in DB, hash:", hash)
	} else {
		fmt.Println("[schema] authority.console v0.9.3 already exists, hash:", hash)
	}
	return nil
}

// ValidateAuthorityInput validates input JSON against the embedded schema.
func ValidateAuthorityInput(input map[string]interface{}) error {
	schemaLoader := gojsonschema.NewBytesLoader(authoritySchema)
	docLoader := gojsonschema.NewGoLoader(input)
	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return fmt.Errorf("validation engine error: %w", err)
	}

	if !result.Valid() {
		var errs []string
		for _, desc := range result.Errors() {
			errs = append(errs, desc.String())
		}
		return errors.New("schema validation failed: " + fmt.Sprint(errs))
	}

	return nil
}

// GetActiveSchema fetches the active schema record from DB.
func GetActiveSchema(ctx context.Context, db *sql.DB) (*SchemaInfo, error) {
	row := db.QueryRowContext(ctx, `
		SELECT id, name, version, json_schema, is_active
		FROM schemas
		WHERE name = 'authority.console' AND is_active = true
		LIMIT 1
	`)
	var s SchemaInfo
	if err := row.Scan(&s.ID, &s.Name, &s.Version, &s.JSON, &s.IsActive); err != nil {
		return nil, err
	}
	s.Hash = calcHash(string(s.JSON))
	return &s, nil
}

// SchemaVerify performs a quick syntax + version check on disk schema files.
func SchemaVerify(schemaFile string) error {
	data, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	var schemaData map[string]any
	if err := json.Unmarshal(data, &schemaData); err != nil {
		return fmt.Errorf("invalid JSON schema %s: %w", schemaFile, err)
	}

	if _, ok := schemaData["version"]; !ok {
		return fmt.Errorf("missing 'version' field in schema %s", schemaFile)
	}
	return nil
}

// calcHash computes a SHA256 hash for integrity checking.
func calcHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return fmt.Sprintf("%x", h.Sum(nil))
}
