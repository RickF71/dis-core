package ledger

import (
	"fmt"
	"path/filepath"
)

// Ledger represents the full DIS ledger system.
type Ledger struct {
	receipts       map[string]Receipt
	schemas        *SchemaRegistry // ← new
	rootPath       string
	schemaBasePath string
	version        string
}

// NewLedger creates and initializes a ledger instance.
func NewLedger(rootPath string, version string) (*Ledger, error) {
	ld := &Ledger{
		receipts:       make(map[string]Receipt),
		rootPath:       rootPath,
		schemaBasePath: filepath.Join(rootPath, "schemas"),
		version:        version,
	}

	// Initialize schema registry.
	ld.schemas = NewSchemaRegistry(ld.schemaBasePath)

	// Auto-register schemas from filesystem.
	fmt.Println("🔍 Scanning for schemas...")
	if err := ld.schemas.AutoRegisterSchemasFromFS("system"); err != nil {
		fmt.Printf("⚠️  Schema auto-registration error: %v\n", err)
	} else {
		fmt.Printf("✅ Schema registry loaded (%d schemas)\n", len(ld.schemas.ListSchemas()))
	}

	return ld, nil
}

// RegisterSchema wraps schema registration for external calls.
func (l *Ledger) RegisterSchema(id, version string, data []byte, by, src string) (*SchemaRecord, error) {
	return l.schemas.RegisterSchema(id, version, data, by, src)
}

// ListSchemas returns all schema metadata.
func (l *Ledger) ListSchemas() []*SchemaRecord {
	return l.schemas.ListSchemas()
}

// DumpSchemas prints the registry in human-readable form.
func (l *Ledger) DumpSchemas() string {
	return l.schemas.DumpSchemaRegistry()
}
