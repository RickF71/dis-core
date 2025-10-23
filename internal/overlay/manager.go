package overlay

import (
	"database/sql"
	"dis-core/internal/util"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	db *sql.DB
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) ImportFromYAML(node map[string]any) error {
	yamlBytes, _ := yaml.Marshal(node)
	return util.ImportYAMLToDB(m.db, "overlays", "overlay.yaml", string(yamlBytes))
}
