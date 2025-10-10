package ledger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// DomainRegistry manages domain-schema bindings.
type DomainRegistry struct {
	mu       sync.RWMutex
	domains  map[string]*DomainRecord
	basePath string
	schemas  *SchemaRegistry
}

// NewDomainRegistry creates a new registry with reference to schema registry.
func NewDomainRegistry(basePath string, schemaReg *SchemaRegistry) *DomainRegistry {
	return &DomainRegistry{
		domains:  make(map[string]*DomainRecord),
		basePath: basePath,
		schemas:  schemaReg,
	}
}

// LoadDomains scans /domains and loads YAML metadata.
func (dr *DomainRegistry) LoadDomains() error {
	return filepath.Walk(dr.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var dom DomainRecord
		if err := yaml.Unmarshal(data, &dom); err != nil {
			fmt.Printf("⚠️  YAML parse failed for %s: %v\n", path, err)
			return nil
		}
		dom.SourcePath = path
		dr.domains[dom.ID] = &dom
		return nil
	})
}

// ValidateBindings ensures each domain’s uses_schemas exist in the schema registry.
func (dr *DomainRegistry) ValidateBindings() {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	for id, d := range dr.domains {
		valid := true
		for _, ref := range d.UsesSchemas {
			key := normalizeSchemaKey(ref)
			if _, ok := dr.schemas.schemas[key]; !ok {
				fmt.Printf("❌ Domain %s references missing schema: %s\n", id, ref)
				valid = false
			}
		}
		d.Validated = valid
		d.CheckedAt = time.Now().UTC()
		if valid {
			fmt.Printf("✅ Domain %s bindings verified (%d schemas)\n", id, len(d.UsesSchemas))
		}
	}
}

// DumpDomainMap prints a summary of domain-schema bindings.
func (dr *DomainRegistry) DumpDomainMap() string {
	dr.mu.RLock()
	defer dr.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString("DIS Domain-Schema Binding Map — v0.8.7\n")
	sb.WriteString("---------------------------------------\n")
	for _, d := range dr.domains {
		sb.WriteString(fmt.Sprintf("Domain: %s (%s)\n", d.ID, d.Version))
		sb.WriteString(fmt.Sprintf("Schemas: %v\n", d.UsesSchemas))
		sb.WriteString(fmt.Sprintf("Validated: %v at %s\n", d.Validated, d.CheckedAt.Format(time.RFC3339)))
		sb.WriteString("---------------------------------------\n")
	}
	return sb.String()
}

func normalizeSchemaKey(ref string) string {
	ref = strings.TrimSpace(ref)
	if strings.Contains(ref, "@") {
		return ref
	}
	if strings.Count(ref, ".v") == 1 {
		parts := strings.Split(ref, ".v")
		return fmt.Sprintf("%s@v%s", parts[0], parts[1])
	}
	return ref
}
