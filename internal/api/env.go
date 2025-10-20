package api

import (
	"os"
	"path/filepath"
)

// getEnv returns ENV[key] or fallback if not set.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// resolveRepoRoot determines the base repo path for domain operations.
func resolveRepoRoot() string {
	return getEnv("DIS_REPO_ROOT", "domains")
}

// resolveSchemasDir determines the schema folder.
func resolveSchemasDir() string {
	return getEnv("DIS_SCHEMAS_DIR", "schemas")
}

// resolveDataDir determines where dynamic DB/canon data lives.
func resolveDataDir() string {
	return getEnv("DIS_DATA_DIR", filepath.Join("data"))
}
