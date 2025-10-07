package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultDomain      string `yaml:"default_domain"`
	DefaultScope       string `yaml:"default_scope"`
	SignatureAlgorithm string `yaml:"signature_algorithm"`
	DatabasePath       string `yaml:"database_path"`
	NonceBytes         int    `yaml:"nonce_bytes"`
	PolicyPath         string `yaml:"policy_path"`
	APIHost            string `yaml:"api_host"`
	APIPort            int    `yaml:"api_port"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil { return nil, err }
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil { return nil, err }
	if c.DefaultDomain == "" { c.DefaultDomain = "domain.null" }
	if c.DefaultScope == "" { c.DefaultScope = "identity.confirm" }
	if c.SignatureAlgorithm == "" { c.SignatureAlgorithm = "sha256" }
	if c.DatabasePath == "" { c.DatabasePath = "data/dis.db" }
	if c.NonceBytes <= 0 { c.NonceBytes = 16 }
	if c.PolicyPath == "" { c.PolicyPath = "policy.yaml" }
	if c.APIHost == "" { c.APIHost = "0.0.0.0" }
	if c.APIPort == 0 { c.APIPort = 8080 }
	return &c, nil
}
