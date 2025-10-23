package domain

import (
	"dis-core/internal/util"

	"gopkg.in/yaml.v3"
)

func (m *Manager) ImportFromYAML(data map[string]any) error {
	yamlBytes, _ := yaml.Marshal(data)
	return util.ImportYAMLToDB(m.db, "domains", "domain.yaml", string(yamlBytes))
}
