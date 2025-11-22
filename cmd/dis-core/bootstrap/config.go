package bootstrap

import (
	"os"

	"dis-core/internal/config"
)

// Config holds all configuration needed for DIS-Core startup
type Config struct {
	Host    string
	Port    string
	Version string
	DSN     string
}

// LoadConfig initializes configuration from environment variables with sensible defaults.
// If DIS_TEST_DB_DSN is set it will be preferred so test runs can target the test DB
// without having to override the other DSN-related env vars.
func LoadConfig() *Config {
	// Prefer explicit test DSN when present
	dsn := os.Getenv("DIS_TEST_DB_DSN")
	if dsn == "" {
		dsn = config.DatabaseURL()
	}

	return &Config{
		Host:    getenvDefault("DIS_API_HOST", "0.0.0.0"),
		Port:    getenvDefault("DIS_API_PORT", "8080"),
		Version: getenvDefault("DIS_VERSION", "v1.0-dev"),
		DSN:     dsn,
	}
}

// Address returns the formatted host:port address for the HTTP server
func (c *Config) Address() string {
	return c.Host + ":" + c.Port
}

// getenvDefault returns the environment variable value or a default if not set
func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
