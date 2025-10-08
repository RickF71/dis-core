package console

import (
	"os"

	"gopkg.in/yaml.v3"
)

type NetworkConfig struct {
	Peers []struct {
		Name       string `yaml:"name"`
		URL        string `yaml:"url"`
		TrustLevel string `yaml:"trust_level"`
	} `yaml:"peers"`
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
