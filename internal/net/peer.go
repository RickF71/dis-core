package net

import (
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// NetworkConfig and config loader (from old peers.go, now in net)
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

// Peer represents another DIS node.
type Peer struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	LastSeen  time.Time `json:"last_seen"`
	Healthy   bool      `json:"healthy"`
	Version   string    `json:"version"`
	LatencyMS int64     `json:"latency_ms"`
}

// PingPeer checks if a peer responds to /api/status.
func PingPeer(address string) (*Peer, error) {
	start := time.Now()
	resp, err := http.Get(address + "/api/status")
	if err != nil {
		return &Peer{Address: address, Healthy: false, LastSeen: time.Now()}, err
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()
	ok := resp.StatusCode == 200
	return &Peer{
		Address:   address,
		LastSeen:  time.Now(),
		Healthy:   ok,
		LatencyMS: latency,
	}, nil
}
