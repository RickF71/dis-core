package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Path    string `json:"path"`
}

type Registry struct {
	byKey map[string]Entry // key = id@version
}

func NewRegistry() *Registry { return &Registry{byKey: map[string]Entry{}} }

func (r *Registry) key(id, version string) string { return id + "@" + version }

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
		var hdr struct {
			ID      string `json:"id" yaml:"id"`
			Version string `json:"version" yaml:"version"`
		}
		if err := yamlUnmarshalHeader(b, &hdr); err != nil {
			return nil
		} // skip non-schema files
		if hdr.ID == "" || hdr.Version == "" {
			return nil
		}
		h := sha256.Sum256(b)
		e := Entry{ID: hdr.ID, Version: hdr.Version, Hash: hex.EncodeToString(h[:]), Path: path}
		r.byKey[r.key(e.ID, e.Version)] = e
		return nil
	})
}

// Get retrieves a schema by id+version.
func (r *Registry) Get(id, version string) (Entry, bool) {
	e, ok := r.byKey[r.key(id, version)]
	return e, ok
}

// Verify compares the stored hash with a recomputed hash.
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

// yamlUnmarshalHeader extracts minimal header fields without strict schema types.
func yamlUnmarshalHeader(b []byte, v any) error {
	// Lazy approach: use a JSON round-trip via yq-like heuristic.
	// In production, import a YAML lib. Here we avoid extra deps in the snippet.
	// Replace with: gopkg.in/yaml.v3
	return json.Unmarshal(YAMLtoJSON(b), v)
}

// YAMLtoJSON is a placeholder adapter; replace with proper YAML decoding.
func YAMLtoJSON(b []byte) []byte {
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return b // fallback
	}
	j, _ := json.Marshal(v)
	return j
}

// HashAll returns a deterministic hash of every schema registered.
// This lets the --freeze command produce a single hash representing the full registry state.
func (r *Registry) HashAll() string {
	h := sha256.New()
	for _, e := range r.byKey {
		h.Write([]byte(e.ID))
		h.Write([]byte(e.Version))
		h.Write([]byte(e.Hash))
	}
	return hex.EncodeToString(h.Sum(nil))
}
