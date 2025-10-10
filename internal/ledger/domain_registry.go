package ledger

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// NowUTC returns the current UTC time in RFC3339 format.

// NowUTC returns the current UTC timestamp.
func NowUTC() time.Time {
	return time.Now().UTC()
}

// LoadDomainsFromFS scans a folder for domain YAML files,
// parses them, validates each schema reference, and records provenance.
func (l *Ledger) LoadDomainsFromFS(path string) ([]*DomainRecord, error) {
	var domains []*DomainRecord

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(p) != ".yaml" {
			return nil
		}

		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		// Temporary structure just to read meta header
		var meta struct {
			Meta struct {
				SchemaID      string `yaml:"schema_id"`
				SchemaVersion string `yaml:"schema_version"`
				DomainID      string `yaml:"domain_id"`
				Description   string `yaml:"description"`
			} `yaml:"meta"`
		}
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return fmt.Errorf("parse error in %s: %v", p, err)
		}

		schemaID := strings.TrimSpace(meta.Meta.SchemaID)
		schemaVer := strings.TrimSpace(meta.Meta.SchemaVersion)
		domainID := strings.TrimSpace(meta.Meta.DomainID)

		if schemaID == "" || schemaVer == "" {
			fmt.Printf("⚠️  Skipping %s — missing meta.schema_id or meta.schema_version\n", p)
			return nil
		}

		var dom DomainRecord
		dom.ID = domainID
		dom.SchemaRef = schemaID
		dom.Version = schemaVer
		dom.SourcePath = p

		// Verify that the referenced schema exists in the registry
		if entry, ok := l.schemas.GetSchema(schemaID, schemaVer); ok {
			dom.Validated = true
			dom.IsBound = true
			dom.CheckedAt = NowUTC()
			dom.SchemaRef = entry.ID
			dom.Version = entry.Version
			fmt.Printf("✅ Domain %s linked to schema %s (%s)\n", domainID, schemaID, schemaVer)
		} else {
			dom.Validated = false
			dom.IsBound = false
			fmt.Printf("⚠️  Domain %s references unknown schema %s (%s)\n", domainID, schemaID, schemaVer)
		}

		domains = append(domains, &dom)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("domain scan failed: %v", err)
	}

	fmt.Printf("✅ Loaded %d domain definitions\n", len(domains))
	return domains, nil
}
