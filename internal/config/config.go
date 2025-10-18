package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config defines runtime configuration for DIS-CORE (PostgreSQL-only).
type Config struct {
	DefaultDomain      string `yaml:"default_domain"`
	DefaultScope       string `yaml:"default_scope"`
	SignatureAlgorithm string `yaml:"signature_algorithm"`
	NonceBytes         int    `yaml:"nonce_bytes"`
	PolicyPath         string `yaml:"policy_path"`
	APIHost            string `yaml:"api_host"`
	APIPort            int    `yaml:"api_port"`
	// RepoRoot allows the server to locate the repository root when resolving
	// domain files. Defaults to "." (current working directory).
	RepoRoot           string `yaml:"repo_root"`

	// PostgreSQL connection string, e.g.:
	// postgres://user:pass@localhost:5432/dis_core?sslmode=disable
	DatabaseDSN string `yaml:"database_dsn"`
}

// Load reads and parses the YAML config file, applying safe defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	// --- Defaults ---
	if c.DefaultDomain == "" {
		c.DefaultDomain = "domain.terra"
	}
	if c.DefaultScope == "" {
		c.DefaultScope = "identity.confirm"
	}
	if c.SignatureAlgorithm == "" {
		c.SignatureAlgorithm = "sha256"
	}
	if c.NonceBytes <= 0 {
		c.NonceBytes = 16
	}
	if c.PolicyPath == "" {
		c.PolicyPath = "policy.yaml"
	}
	if c.APIHost == "" {
		c.APIHost = "0.0.0.0"
	}
	if c.APIPort == 0 {
		c.APIPort = 8080
	}

	if c.RepoRoot == "" {
		c.RepoRoot = "."
	}

	// Allow DSN from env var if not in YAML
	if c.DatabaseDSN == "" {
		if env := os.Getenv("DIS_DB_DSN"); env != "" {
			c.DatabaseDSN = env
		} else {
			c.DatabaseDSN = "postgres://dis_user:card567@localhost:5432/dis_core?sslmode=disable"
		}
	}

	return &c, nil
}
