package domain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Domain represents the minimal domain manifest fields we expose.
type Domain struct {
	Code        string   `yaml:"code" json:"code"`
	Name        string   `yaml:"name" json:"name"`
	Seat        string   `yaml:"seat" json:"seat"`
	Lineage     []string `yaml:"lineage" json:"lineage"`
	Population  int64    `yaml:"population" json:"population"`
	Description string   `yaml:"description" json:"description"`
}

// fileExists reports whether path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// LookupByCode accepts a 3-letter ISO code (case-insensitive). It searches
// for /domains/countries/{iso}/{iso}.yaml first, then falls back to
// /domains/terra/{iso}.yaml. The function returns a parsed *Domain.
// Errors are returned with clear 400 or 404 prefixes for callers to map to
// HTTP responses if desired.
func LookupByCode(code string, repoRoot string) (*Domain, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("400 Bad Request: missing domain code")
	}

	iso := strings.ToLower(strings.TrimSpace(code))
	// Expect ISO A3 (3 letters)
	if len(iso) != 3 {
		return nil, fmt.Errorf("400 Bad Request: invalid ISO code (expect 3 letters): %q", code)
	}

	// canonical path: ./domains/countries/<iso>/domain.<iso>.yaml
	fname := fmt.Sprintf("domain.%s.yaml", iso)
	found := filepath.Join(repoRoot, "domains", "countries", iso, fname)

	if !fileExists(found) {
		return nil, fmt.Errorf("404 Not Found: domain for code %q not found at canonical path %s", code, found)
	}

	b, err := os.ReadFile(found)
	if err != nil {
		return nil, fmt.Errorf("400 Bad Request: failed to read domain file %s: %v", found, err)
	}

	var d Domain
	if err := yaml.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("400 Bad Request: failed to parse domain YAML %s: %v", found, err)
	}

	// Normalize code field to uppercase ISO
	d.Code = strings.ToUpper(d.Code)
	return &d, nil
}
