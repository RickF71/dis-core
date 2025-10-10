package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Entry represents a single registered schema.
type Entry struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Path    string `json:"path"`
}

// Registry holds all loaded schemas in memory.
type Registry struct {
	byKey map[string]Entry // key = id@version
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry { return &Registry{byKey: map[string]Entry{}} }

// key builds the canonical lookup key for a schema.
func (r *Registry) key(id, version string) string {
	return fmt.Sprintf("%s@%s", strings.TrimSpace(id), strings.TrimSpace(version))
}

// LoadDir walks a directory and registers any YAML matching a DIS schema header.
func (r *Registry) LoadDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".yaml") && !strings.HasSuffix(d.Name(), ".yml") {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Expect "meta.schema_id" and "meta.schema_version"
		var hdr struct {
			Meta struct {
				SchemaID      string `yaml:"schema_id"`
				SchemaVersion string `yaml:"schema_version"`
			} `yaml:"meta"`
		}
		if err := yaml.Unmarshal(b, &hdr); err != nil {
			return err
		}

		if hdr.Meta.SchemaID == "" || hdr.Meta.SchemaVersion == "" {
			return nil // skip non-schema YAMLs
		}

		// Strict version enforcement: must start with 'v' and contain '.'
		if !strings.HasPrefix(hdr.Meta.SchemaVersion, "v") || !strings.Contains(hdr.Meta.SchemaVersion, ".") {
			return fmt.Errorf("❌ invalid schema version in %s: '%s' (must be like v1.0)",
				path, hdr.Meta.SchemaVersion)
		}

		// Compute content hash
		h := sha256.Sum256(b)
		hashHex := hex.EncodeToString(h[:])

		e := Entry{
			ID:      hdr.Meta.SchemaID,
			Version: hdr.Meta.SchemaVersion,
			Hash:    hashHex,
			Path:    path,
		}
		r.byKey[r.key(e.ID, e.Version)] = e

		fmt.Printf("📜 Registered schema: %s (%s)\n", e.ID, e.Version)
		return nil
	})
}

// Get retrieves a schema by id + version.
func (r *Registry) Get(id, version string) (Entry, bool) {
	e, ok := r.byKey[r.key(id, version)]
	return e, ok
}

// Verify recomputes and checks a schema’s hash.
func (r *Registry) Verify(id, version string) error {
	e, ok := r.Get(id, version)
	if !ok {
		return fmt.Errorf("schema not found: %s@%s", id, version)
	}
	b, err := os.ReadFile(e.Path)
	if err != nil {
		return err
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != e.Hash {
		return errors.New("schema hash mismatch")
	}
	return nil
}

// HashAll returns a deterministic hash of all schemas.
func (r *Registry) HashAll() string {
	h := sha256.New()
	keys := make([]string, 0, len(r.byKey))
	for k := range r.byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		e := r.byKey[k]
		h.Write([]byte(e.ID))
		h.Write([]byte(e.Version))
		h.Write([]byte(e.Hash))
	}
	return hex.EncodeToString(h.Sum(nil))
}
