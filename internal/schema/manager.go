package schema

import "database/sql"

type Manager struct{ db *sql.DB }

func NewManager(db *sql.DB) *Manager { return &Manager{db: db} }

func (m *Manager) ImportFromYAML(node map[string]any) error {
	// TODO: parse & insert schema YAML into DB
	return nil
}
