package schema

import (
	"database/sql"
	"dis-core/internal/util"
)

// DefaultSchemaDir defines the canonical location for DIS schema files.
// All schema managers and loaders should build paths relative to this root.
var DefaultSchemaDir = "schemas"

type Manager struct{ db *sql.DB }

func NewManager(db *sql.DB) *Manager { return &Manager{db: db} }

func (m *Manager) ImportFromYAML(node map[string]any) error {
	_, _, err := util.ImportYAML(DefaultSchemaDir, "schema.yaml")
	return err
}
