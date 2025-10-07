package version

import (
	"crypto/sha256"
	"encoding/hex"
	"os"

	"gopkg.in/yaml.v3"
)

type VersionInfo struct {
	DISCore     string `yaml:"dis-core"`
	DISPersonal string `yaml:"dis-personal"`
	Status      string `yaml:"status"`
	Notes       string `yaml:"notes"`
}

// Load reads VERSION.yaml from disk and returns structured info.
func Load(path string) (*VersionInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v VersionInfo
	err = yaml.Unmarshal(data, &v)
	return &v, err
}

// CoreChecksum computes a SHA256 hash of the given core schema file.
func CoreChecksum(corePath string) (string, error) {
	data, err := os.ReadFile(corePath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
