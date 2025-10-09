package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"dis-core/internal/schema"

	"gopkg.in/yaml.v3"
)

// DomainDoc represents a parsed domain definition (from /domains/*.yaml)
type DomainDoc struct {
	Meta struct {
		SchemaID      string `json:"schema_id" yaml:"schema_id"`
		SchemaVersion string `json:"schema_version" yaml:"schema_version"`
		SchemaHash    string `json:"schema_hash" yaml:"schema_hash"`
		Name          string `json:"name" yaml:"name"`
		UUID          string `json:"uuid" yaml:"uuid"`
		Parent        string `json:"parent_domain_ref" yaml:"parent_domain_ref"`
	} `json:"meta" yaml:"meta"`
}

// LoadAndValidate reads a domain YAML and validates its schema binding via registry.
func LoadAndValidate(path string, reg *schema.Registry) (*DomainDoc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var d DomainDoc
	// ✅ Proper YAML→JSON conversion before unmarshal
	if err := json.Unmarshal(YAMLtoJSON(b), &d); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if d.Meta.SchemaID == "" || d.Meta.SchemaVersion == "" {
		return nil, fmt.Errorf("missing meta.schema_id/version in %s", filepath.Base(path))
	}

	// Verify schema hash
	if err := reg.Verify(d.Meta.SchemaID, d.Meta.SchemaVersion); err != nil {
		return nil, fmt.Errorf("schema verify failed for %s@%s: %w",
			d.Meta.SchemaID, d.Meta.SchemaVersion, err)
	}

	return &d, nil
}

// YAMLtoJSON safely converts YAML to JSON for json.Unmarshal compatibility
func YAMLtoJSON(b []byte) []byte {
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return b // fallback if invalid YAML
	}
	j, err := json.Marshal(v)
	if err != nil {
		return b
	}
	return j
}
