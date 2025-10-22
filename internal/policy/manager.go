package policy

import "database/sql"

type Manager struct {
	db *sql.DB
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) ImportFromYAML(node map[string]any) error {
	// TODO: parse and insert policy YAML into DB
	return nil
}
