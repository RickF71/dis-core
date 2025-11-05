package reconcile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dis-core/internal/domain"
	"dis-core/internal/schema"
)

// --- Structs (mirror API warnings) ---

type SchemaWarning struct {
	ID         int        `json:"id,omitempty"`
	FilePath   string     `json:"file"`
	Domain     string     `json:"domain"`
	SchemaType string     `json:"schema_type"`
	Details    string     `json:"details"`
	Resolved   bool       `json:"resolved"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

type DomainWarning struct {
	ID         int        `json:"id,omitempty"`
	Domain     string     `json:"domain"`
	Parent     string     `json:"parent"`
	Details    string     `json:"details"`
	Resolved   bool       `json:"resolved"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// --- Validators ---

// ValidateDomains scans /domains/* YAML files for lineage or schema issues.
func ValidateDomains(repoRoot string, reg *schema.Registry) ([]DomainWarning, error) {
	var warnings []DomainWarning

	err := filepath.Walk(filepath.Join(repoRoot, "domains"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasPrefix(info.Name(), "domain.") || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		code := strings.TrimSuffix(strings.TrimPrefix(info.Name(), "domain."), ".yaml")
		d, err := domain.LookupByCode(code, repoRoot)
		if err != nil {
			warnings = append(warnings, DomainWarning{
				Domain:    code,
				Details:   fmt.Sprintf("failed to parse YAML: %v", err),
				Resolved:  false,
				CreatedAt: time.Now(),
			})
			return nil
		}

		if len(d.Lineage) == 0 {
			warnings = append(warnings, DomainWarning{
				Domain:    d.Code,
				Details:   "missing lineage (no parent domain listed)",
				Resolved:  false,
				CreatedAt: time.Now(),
			})
		}

		// Check schema existence
		schemaName := fmt.Sprintf("schema.%s", strings.ToLower(d.Code))
		if _, ok := reg.Get(schemaName, "v1"); !ok {
			warnings = append(warnings, DomainWarning{
				Domain:    d.Code,
				Parent:    last(d.Lineage),
				Details:   fmt.Sprintf("schema %s not found (using default)", schemaName),
				Resolved:  false,
				CreatedAt: time.Now(),
			})
		}

		return nil
	})

	return warnings, err
}

// ValidateSchemas emits warnings if registry entries are missing metadata.
func ValidateSchemas(reg *schema.Registry) ([]SchemaWarning, error) {
	// Since *schema.Registry doesn’t expose All(), we can’t enumerate directly yet.
	// For now, return an empty list (or a placeholder warning).
	var warnings []SchemaWarning

	warnings = append(warnings, SchemaWarning{
		FilePath:   "(registry)",
		Domain:     "(n/a)",
		SchemaType: "(summary)",
		Details:    "schema registry enumeration not yet implemented",
		Resolved:   true,
		CreatedAt:  time.Now(),
	})

	return warnings, nil
}

// helper
func last(lineage []string) string {
	if len(lineage) == 0 {
		return ""
	}
	return lineage[len(lineage)-1]
}
