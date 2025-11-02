package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Reconcile writes a bootstrapable .disyaml and syncs type info to DB
func Reconcile(db *sql.DB, id string, blob map[string]any) (map[string]any, error) {
	yamlData, err := yaml.Marshal(blob)
	if err != nil {
		return nil, fmt.Errorf("marshal bootstrapable yaml: %w", err)
	}

	baseDir := "/opt/dis/bootstrap"
	target := filepath.Join(baseDir, fmt.Sprintf("%s_bootstrap.disyaml", id))

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}
	if err := os.WriteFile(target, yamlData, 0o644); err != nil {
		return nil, fmt.Errorf("write yaml: %w", err)
	}

	typeVal := ""
	if t, ok := blob["type"].(string); ok {
		typeVal = t
	}

	jsonBytes, _ := json.Marshal(blob)
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO bootstrap (id, type, content, source_file)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
			SET type = EXCLUDED.type,
			    content = EXCLUDED.content,
			    source_file = EXCLUDED.source_file,
			    updated_at = NOW()
	`, id, typeVal, jsonBytes, target)
	if err != nil {
		return nil, fmt.Errorf("db upsert: %w", err)
	}

	return map[string]any{
		"status": "reconciled",
		"id":     id,
		"type":   typeVal,
		"path":   target,
	}, nil
}
