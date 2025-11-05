package config

import (
	"fmt"
	"os"
)

// DatabaseURL builds the DSN with proper precedence:
// 1) Config.DatabaseDSN (config.yaml or loaded config)
// 2) DIS_DB_DSN env var
// 3) constructed DSN from component env vars
// 4) fallback local dev default
func (c *Config) DatabaseURL() string {
	// #1 Explicit config value wins
	if c.DatabaseDSN != "" {
		return c.DatabaseDSN
	}

	// #2 Env override
	if env := os.Getenv("DIS_DB_DSN"); env != "" {
		return env
	}

	// #3 Construct from ENV pieces or defaults
	user := os.Getenv("DIS_DB_USER")
	if user == "" {
		user = "dis_user"
	}

	pass := os.Getenv("DIS_DB_PASSWORD")
	if pass == "" {
		pass = "card567"
	}

	host := os.Getenv("DIS_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DIS_DB_PORT")
	if port == "" {
		port = "5432"
	}

	name := os.Getenv("DIS_DB_NAME")
	if name == "" {
		name = "dis"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, name,
	)
}
