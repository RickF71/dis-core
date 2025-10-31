package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// ImportYAML is the canonical helper for reading and hashing YAML files.
// It reads the YAML file at the given path, returns its content and SHA256 hash.
//
// Returns: content []byte, hash string, error
func ImportYAML(baseDir, filename string) ([]byte, string, error) {
	fullPath := filepath.Join(baseDir, filename)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read YAML file %s: %w", fullPath, err)
	}

	hash := sha256.Sum256(data)
	return data, hex.EncodeToString(hash[:]), nil
}
