package canon

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// ImportYAMLToDB parses YAML, converts it to JSON, and inserts it into a given table.
func ImportYAMLToDB(db *sql.DB, table string, filename string, yamlContent string) error {
	// Parse YAML into a generic map
	var node map[string]any
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		return fmt.Errorf("yaml parse error: %v", err)
	}

	// Convert to JSON for storage
	jsonData, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}

	// Extract name if present
	name := filename
	if meta, ok := node["meta"].(map[string]any); ok {
		if val, ok := meta["name"].(string); ok {
			name = val
		}
	}

	// Upsert into the target table
	_, err = db.Exec(fmt.Sprintf(`
		INSERT INTO %s (name, data)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET data = EXCLUDED.data
	`, table), name, jsonData)
	if err != nil {
		return fmt.Errorf("db insert error: %v", err)
	}

	return nil
}
