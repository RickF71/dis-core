package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SchemaRecord captures metadata about a DIS schema.
type SchemaRecord struct {
	ID           string    `json:"id" yaml:"id"`
	Version      string    `json:"version" yaml:"version"`
	Hash         string    `json:"hash" yaml:"hash"`
	Registered   time.Time `json:"registered"`
	RegisteredBy string    `json:"registered_by"`
	SourcePath   string    `json:"source_path"`
}

// SchemaRegistry holds all registered schema metadata.
type SchemaRegistry struct {
	mu       sync.RWMutex
	schemas  map[string]*SchemaRecord // keyed by ID
	basePath string                   // seed folder, usually ./schemas
}

// NewSchemaRegistry creates a new in-memory registry.
func NewSchemaRegistry(basePath string) *SchemaRegistry {
	return &SchemaRegistry{
		schemas:  make(map[string]*SchemaRecord),
		basePath: basePath,
	}
}

// RegisterSchema adds or updates a schema record in the registry.
func (sr *SchemaRegistry) RegisterSchema(id, version string, data []byte, registeredBy string, sourcePath string) (*SchemaRecord, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])
	key := fmt.Sprintf("%s@%s", id, version)

	rec := &SchemaRecord{
		ID:           id,
		Version:      version,
		Hash:         hashStr,
		Registered:   time.Now().UTC(),
		RegisteredBy: registeredBy,
		SourcePath:   sourcePath,
	}
	sr.schemas[key] = rec
	return rec, nil
}

func (sr *SchemaRegistry) RegisterCoreSchemas() error {
	_, err1 := sr.RegisterSchema("value_receipt", "v1", nil, "system", "/domains/dis/schemas/core/value_receipt.v1.yaml")
	_, err2 := sr.RegisterSchema("lifepush_substrate_structure", "v1", nil, "system", "/domains/terra/schemas/lifepush_substrate_structure.v1.yaml")
	if err1 != nil {
		return err1
	}
	return err2
}

// ListSchemas returns a slice of all registered schema records.
func (sr *SchemaRegistry) ListSchemas() []*SchemaRecord {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	out := make([]*SchemaRecord, 0, len(sr.schemas))
	for _, s := range sr.schemas {
		out = append(out, s)
	}
	return out
}

// GetSchema fetches a schema record by id and version.
func (sr *SchemaRegistry) GetSchema(id, version string) (*SchemaRecord, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	key := fmt.Sprintf("%s@%s", id, version)
	rec, ok := sr.schemas[key]
	return rec, ok
}

// AutoRegisterSchemasFromFS scans the schema directory and registers unseen files.
func (sr *SchemaRegistry) AutoRegisterSchemasFromFS(registeredBy string) error {
	return filepath.WalkDir(sr.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".yaml") && !strings.HasSuffix(d.Name(), ".yml") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		id, version := extractSchemaMeta(path)
		key := fmt.Sprintf("%s@%s", id, version)

		sr.mu.RLock()
		_, exists := sr.schemas[key]
		sr.mu.RUnlock()
		if exists {
			return nil // already registered
		}

		_, regErr := sr.RegisterSchema(id, version, data, registeredBy, path)
		if regErr != nil {
			return regErr
		}
		fmt.Printf("📜 Registered schema: %s (%s)\n", id, version)
		return nil
	})
}

// DumpSchemaRegistry prints a human-readable summary of all schemas.
func (sr *SchemaRegistry) DumpSchemaRegistry() string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString("DIS Schema Registry — v0.8.7\n")
	sb.WriteString("---------------------------------------\n")
	for _, s := range sr.schemas {
		sb.WriteString(fmt.Sprintf("ID: %s\n", s.ID))
		sb.WriteString(fmt.Sprintf("Version: %s\n", s.Version))
		sb.WriteString(fmt.Sprintf("Hash: %s\n", s.Hash))
		sb.WriteString(fmt.Sprintf("Registered: %s\n", s.Registered.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("RegisteredBy: %s\n", s.RegisteredBy))
		sb.WriteString(fmt.Sprintf("Source: %s\n", s.SourcePath))
		sb.WriteString("---------------------------------------\n")
	}
	return sb.String()
}

// extractSchemaMeta tries to parse id/version from filename (fallback).
func extractSchemaMeta(path string) (id, version string) {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	// convention: name like dis_consent.v0.1
	parts := strings.Split(name, ".v")
	if len(parts) == 2 {
		id = parts[0]
		version = "v" + parts[1]
	} else {
		id = name
		version = "v0.0"
	}
	return
}

// SaveRegistry writes the current schema registry to disk as JSON.
func (sr *SchemaRegistry) SaveRegistry(snapshotPath string) error {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	data, err := json.MarshalIndent(sr.schemas, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(snapshotPath, data, 0644)
}

// LoadRegistry loads a JSON snapshot of the registry from disk.
func (sr *SchemaRegistry) LoadRegistry(snapshotPath string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		// No snapshot yet; skip without failing.
		return nil
	}

	var restored map[string]*SchemaRecord
	if err := json.Unmarshal(data, &restored); err != nil {
		return err
	}

	if sr.schemas == nil {
		sr.schemas = make(map[string]*SchemaRecord)
	}
	for k, v := range restored {
		sr.schemas[k] = v
	}

	fmt.Printf("💾 Loaded schema registry snapshot (%d entries)\n", len(sr.schemas))
	return nil
}

func (sr *SchemaRegistry) Init() {
	if sr.schemas == nil {
		sr.schemas = make(map[string]*SchemaRecord)
	}
}
