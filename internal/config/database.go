package config

import (
	"fmt"
	"os"
)

// DatabaseURL builds the Postgres DSN using environment variables,
// with sensible local defaults for development.
func DatabaseURL() string {
	// 1) Single env var override
	if env := os.Getenv("DIS_DB_DSN"); env != "" {
		return env
	}

	// 2) Compose from parts or defaults
	user := getenvDefault("DIS_DB_USER", "dis_user")
	pass := getenvDefault("DIS_DB_PASSWORD", "card567")
	host := getenvDefault("DIS_DB_HOST", "localhost")
	port := getenvDefault("DIS_DB_PORT", "5432")
	name := getenvDefault("DIS_DB_NAME", "dis")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, name)
}

// getenvDefault returns the env value or a fallback default.
func getenvDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
