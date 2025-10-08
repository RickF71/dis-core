package console

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SaveState writes the console’s current state (metadata + actions) to disk.
func (c *Console) SaveState() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	dir := "versions/v0.6/console"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, c.ID+"_state.json")
	return os.WriteFile(filename, data, 0644)
}

// LoadState reads a previously saved console state from disk.
func LoadState(filePath string) (*Console, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var c Console
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
