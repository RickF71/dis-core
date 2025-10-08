package network

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Peer struct {
	Name         string `yaml:"name"`
	URL          string `yaml:"url"`
	PublicKeyB64 string `yaml:"public_key_b64"`
	TrustLevel   string `yaml:"trust_level"`
	LastSeen     string `yaml:"last_seen"`
	Notes        string `yaml:"notes"`
}

type NetworkConfig struct {
	Version     string                       `yaml:"version"`
	Codename    string                       `yaml:"codename"`
	Peers       []Peer                       `yaml:"peers"`
	TrustLevels map[string]map[string]string `yaml:"trust_levels"`
}

func LoadNetworkConfig(path string) (*NetworkConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg NetworkConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
