package ledger

import (
	"dis-core/internal/schema"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// NowUTC returns the current UTC timestamp.
func NowUTC() time.Time {
	return time.Now().UTC()
}

// validDISName enforces DIS filename convention: <type>.<subject>[.<version>].disyaml
var validDISName = regexp.MustCompile(`^[a-z]+(\.[a-z0-9_-]+){1,2}\.disyaml$`)

// BootstrapDomains imports all domain .disyaml files, validates them against schemas,
// stores them canonically, and records receipts for each successful import.
func (l *Ledger) BootstrapDomains(reg *schema.Registry, dir string) error {
	fmt.Println("📂 Scanning DIS domains in:", dir)

	domains, err := l.LoadDomainsFromFS(dir, reg)
	if err != nil {
		return fmt.Errorf("failed to load domains: %w", err)
	}

	var stored, failed int
	for _, dom := range domains {
		if !dom.Validated {
			fmt.Printf("⚠️  Skipping unvalidated domain: %s\n", dom.ID)
			continue
		}

		if err := l.StoreCanon(dom); err != nil {
			failed++
			fmt.Printf("❌ Failed to store %s: %v\n", dom.ID, err)
			continue
		}
		stored++
		fmt.Printf("💾 Stored domain canon: %s\n", dom.ID)

		msg := fmt.Sprintf("Imported domain %s from %s", dom.ID, dom.SourcePath)
		if _, err := l.RecordImport(dom.SourcePath, msg); err != nil {
			fmt.Printf("⚠️  Failed to record import receipt for %s: %v\n", dom.ID, err)
		}
	}

	fmt.Printf("✅ Imported %d domains (%d failed)\n", stored, failed)
	return nil
}

// LoadDomainsFromFS scans a folder for valid *.disyaml domain files,
// parses them, validates schema references, and records provenance.
func (l *Ledger) LoadDomainsFromFS(path string, reg *schema.Registry) ([]*DomainRecord, error) {
	var domains []*DomainRecord

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := filepath.Base(p)
		if !validDISName.MatchString(name) {
			// silently skip anything that’s not a valid DIS file
			return nil
		}
		if !strings.HasPrefix(name, "domain.") {
			// only process domain.*.disyaml here
			return nil
		}

		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}

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
			fmt.Printf("⚠️  Skipping %s — missing meta.schema_id or meta.schema_version\n", name)
			return nil
		}

		dom := DomainRecord{
			ID:         domainID,
			SchemaRef:  schemaID,
			Version:    schemaVer,
			SourcePath: p,
		}

		if entry, ok := reg.Get(schemaID, schemaVer); ok {
			dom.Validated = true
			dom.IsBound = true
			dom.CheckedAt = NowUTC()
			dom.SchemaRef = entry.ID
			dom.Version = entry.Version
			fmt.Printf("✅ Domain %s linked to schema %s (%s)\n", dom.ID, schemaID, schemaVer)
		} else {
			dom.Validated = false
			dom.IsBound = false
			fmt.Printf("⚠️  Domain %s references unknown schema %s (%s)\n", dom.ID, schemaID, schemaVer)
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
