// internal/authority/schema.go
package authority

import (
	"encoding/json"
	"os"
)

// AuthoritySchema defines the in-memory Authority Console schema.
type AuthoritySchema struct {
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Schema      map[string]interface{} `json:"schema"`
}

// LoadSchema reads schema.json from a provided path (read-only).
func LoadSchema(path string) (*AuthoritySchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s AuthoritySchema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
