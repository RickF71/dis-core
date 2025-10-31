package schema

import (
	"database/sql"
	"dis-core/internal/util"
)

type Manager struct{ db *sql.DB }

func NewManager(db *sql.DB) *Manager { return &Manager{db: db} }

func (m *Manager) ImportFromYAML(node map[string]any) error {
	_, _, err := util.ImportYAML(DefaultSchemaDir, "schema.yaml")
	return err
}
