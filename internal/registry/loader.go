package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Canon struct {
	Path string
	Kind string
	Blob []byte
}

var canons = map[string]*Canon{} // key = absolute path

// Semi-strict filename rule: <kind>.<subject>[.<version>].disyaml
var validName = regexp.MustCompile(`^[a-z]+(\.[a-z0-9_-]+){1,2}\.disyaml$`)

// LoadRegistry walks the repo root and registers all valid *.disyaml files.
// Only files that match the DIS naming pattern are imported.
// LoadRegistry walks the repo root and registers all valid *.disyaml files.
// Schemas are loaded first, then all other artifacts (domains, events, etc.).
func LoadRegistry(root string) error {
	glob := filepath.Join(root, "**", "*.disyaml")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return fmt.Errorf("scan *.disyaml: %w", err)
	}

	// Separate schemas first
	var schemas, others []string
	for _, p := range matches {
		name := filepath.Base(p)
		if !validName.MatchString(name) {
			fmt.Printf("⚠️  Skipping invalid name: %s\n", name)
			continue
		}
		if strings.HasPrefix(name, "schema.") || strings.Contains(name, ".schema.") {
			schemas = append(schemas, p)
		} else {
			others = append(others, p)
		}
	}

	loaded := 0

	// Phase 1 — load all schemas
	for _, p := range schemas {
		data, err := os.ReadFile(p)
		if err != nil {
			fmt.Printf("❌  Failed to read schema %s: %v\n", p, err)
			continue
		}
		a := abs(p)
		canons[a] = &Canon{Path: a, Kind: "schema", Blob: data}
		fmt.Printf("📘 Schema    %s\n", filepath.Base(p))
		loaded++
	}

	// Phase 2 — load all other domain artifacts
	for _, p := range others {
		data, err := os.ReadFile(p)
		if err != nil {
			fmt.Printf("❌  Failed to read %s: %v\n", p, err)
			continue
		}
		a := abs(p)
		kind := kindFromFilename(filepath.Base(p))
		canons[a] = &Canon{Path: a, Kind: kind, Blob: data}
		fmt.Printf("💾 Loaded %-8s %s\n", kind, filepath.Base(p))
		loaded++
	}

	fmt.Printf("✅ Registry loaded: %d total (%d schemas, %d others)\n",
		loaded, len(schemas), len(others))
	return nil
}

// kindFromFilename extracts the prefix before the first dot.
// Example: "domain.terra.disyaml" → "domain"
func kindFromFilename(name string) string {
	if i := strings.Index(name, "."); i > 0 {
		return name[:i]
	}
	return "unknown"
}

func abs(p string) string {
	a, _ := filepath.Abs(p)
	return a
}

// ReplaceByPath updates a registered definition and writes it back to disk.
func ReplaceByPath(targetPath string, newBlob []byte, repoRoot string) error {
	absTarget := targetPath
	if !filepath.IsAbs(targetPath) {
		absTarget = filepath.Clean(filepath.Join(repoRoot, targetPath))
	}

	entry, ok := canons[absTarget]
	if !ok {
		return fmt.Errorf("definition not loaded: %s", absTarget)
	}
	entry.Blob = newBlob

	if err := os.WriteFile(absTarget, newBlob, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", absTarget, err)
	}
	fmt.Printf("🔄 Updated %s\n", filepath.Base(absTarget))
	return nil
}
