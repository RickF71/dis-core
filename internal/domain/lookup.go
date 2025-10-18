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

	// build candidate paths
	// 1) /domains/countries/{iso}/{iso}.yaml
	p1 := filepath.Join(repoRoot, "domains", "countries", iso, fmt.Sprintf("%s.yaml", iso))
	// also accept legacy filename domain.{iso}.yaml for robustness
	p1alt := filepath.Join(repoRoot, "domains", "countries", iso, fmt.Sprintf("domain.%s.yaml", iso))

	// 2) /domains/terra/{iso}.yaml
	p2 := filepath.Join(repoRoot, "domains", "terra", fmt.Sprintf("%s.yaml", iso))
	p2alt := filepath.Join(repoRoot, "domains", "terra", fmt.Sprintf("domain.%s.yaml", iso))

	candidates := []string{p1, p1alt, p2, p2alt}
	var found string
	for _, p := range candidates {
		if fileExists(p) {
			found = p
			break
		}
	}

	if found == "" {
		return nil, fmt.Errorf("404 Not Found: domain for code %q not found", code)
	}

	b, err := os.ReadFile(found)
	if err != nil {
		return nil, fmt.Errorf("400 Bad Request: failed to read domain file %s: %v", found, err)
	}

	var d Domain
	if err := yaml.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("400 Bad Request: failed to parse domain file %s: %v", found, err)
	}

	// Normalize code field to uppercase ISO
	d.Code = strings.ToUpper(d.Code)
	return &d, nil
}
