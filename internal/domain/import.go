package domain

import (
	"dis-core/internal/util"
)

func (m *Manager) ImportFromYAML(data map[string]any) error {
	_, _, err := util.ImportYAML("domains", "domain.yaml")
	return err
}
