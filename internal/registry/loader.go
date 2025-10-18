package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

type Canon struct {
	Path string
	Kind string
	Blob []byte
}

var canons = map[string]*Canon{} // key = absolute path

// LoadAllDomains walks /domains/* and registers canons by absolute path
func LoadAllDomains(root string) error {
	glob := filepath.Join(root, "domains", "*", "*_canon.yaml")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	for _, p := range matches {
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		canons[abs(p)] = &Canon{
			Path: abs(p),
			Kind: detectKind(data), // TODO: parse YAML "kind:"
			Blob: data,
		}
	}
	return nil
}

func abs(p string) string {
	a, _ := filepath.Abs(p)
	return a
}

func detectKind(_ []byte) string { return "unknown" }

// ReplaceByPath swaps the canonical blob using an absolute/relative path
func ReplaceByPath(targetPath string, newBlob []byte, repoRoot string) error {
	absTarget := targetPath
	if !filepath.IsAbs(targetPath) {
		absTarget = filepath.Clean(filepath.Join(repoRoot, "domains", "dis", "foundation", targetPath))
		// If not relative to foundation, try repo root:
		if _, err := os.Stat(absTarget); err != nil {
			absTarget = filepath.Clean(filepath.Join(repoRoot, targetPath))
		}
	}

	entry, ok := canons[absTarget]
	if !ok {
		return fmt.Errorf("canon not loaded: %s", absTarget)
	}
	entry.Blob = newBlob
	return os.WriteFile(absTarget, newBlob, 0644)
}
