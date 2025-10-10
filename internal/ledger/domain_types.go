package ledger

import "time"

// DomainRecord represents both the static definition (from YAML)
// and the dynamic runtime state (binding, validation, etc.)
type DomainRecord struct {
	// --- YAML-defined fields ---
	ID          string `yaml:"id"`          // unique domain ID, e.g., domain.terra
	Version     string `yaml:"version"`     // semantic version
	SchemaRef   string `yaml:"schema_ref"`  // ID of governing schema
	Description string `yaml:"description"` // human-readable summary

	// --- Optional YAML extensions ---
	UsesSchemas []string `yaml:"uses_schemas,omitempty"` // domains may depend on multiple schemas
	Authority   string   `yaml:"authority,omitempty"`    // linked controlling seat or entity

	// --- Runtime-only fields ---
	SourcePath string    `yaml:"-"` // file path loaded from
	Validated  bool      `yaml:"-"` // whether structure was successfully validated
	CheckedAt  time.Time `yaml:"-"` // last validation timestamp
	IsBound    bool      `yaml:"-"` // whether schema binding succeeded
}
