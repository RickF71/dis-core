package policy

import (
	"fmt"
	"os"
	"path/filepath"
)

func LoadDomainModules(domainID string) (map[string]string, error) {
	// Check domain-specific path first
	base := filepath.Join("domains", domainID, "policy")
	files := []string{"gates.rego", "risk.rego", "freeze.rego"}
	modules := map[string]string{}

	for _, f := range files {
		path := filepath.Join(base, f)
		data, err := os.ReadFile(path)
		if err != nil {
			// fallback to internal default
			defaultPath := filepath.Join("internal/policy", f)
			data, derr := os.ReadFile(defaultPath)
			if derr != nil {
				return nil, fmt.Errorf("failed to read %s or fallback: %w", f, derr)
			}
			modules[f] = string(data)
			continue
		}
		modules[f] = string(data)
	}
	return modules, nil
}
