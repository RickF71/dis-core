package registry

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// BootstrapImportYAML scans /disyaml for YAML files, parses them,
// and loads them into the "bootstrap" DB table.
// This layer is *not canon* — it’s a working reflection of the filesystem.
func BootstrapImportYAML(root string, db *sql.DB) error {
	disyaml := filepath.Join(root, "disyaml")
	if _, err := os.Stat(disyaml); os.IsNotExist(err) {
		return fmt.Errorf("missing disyaml directory at %s", disyaml)
	}

	// ensure bootstrap table exists
	if err := ensureBootstrapTable(db); err != nil {
		return fmt.Errorf("ensure table: %w", err)
	}

	count := 0

	err := filepath.Walk(disyaml, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("⚠️  read error: %s (%v)\n", path, err)
			return nil
		}

		var node map[string]any
		if err := yaml.Unmarshal(data, &node); err != nil {
			fmt.Printf("⚠️  invalid YAML: %s (%v)\n", path, err)
			return nil
		}

		id := detectBootstrapID(path, node)

		if err := storeBootstrap(db, id, node, path); err != nil {
			fmt.Printf("⚠️  DB insert failed: %s (%v)\n", path, err)
			return nil
		}

		fmt.Printf("✅ Imported %s\n", path)
		count++
		return nil
	})

	if err != nil {
		return fmt.Errorf("walk disyaml: %w", err)
	}

	fmt.Printf("🟢 Bootstrap import complete — %d YAML objects loaded\n", count)
	return nil
}

// detectBootstrapID tries to find a stable ID for the YAML object.
func detectBootstrapID(path string, node map[string]any) string {
	if id, ok := node["id"].(string); ok && id != "" {
		return id
	}
	if name, ok := node["domain"].(string); ok && name != "" {
		return name
	}
	return strings.TrimSuffix(filepath.Base(path), ".yaml")
}

// ensureBootstrapTable creates the table if it doesn’t exist.
func ensureBootstrapTable(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS bootstrap (
		id TEXT PRIMARY KEY,
		type TEXT,
		content JSONB,
		source_file TEXT,
		hash TEXT,
		imported_at TIMESTAMPTZ DEFAULT NOW()
	);`
	_, err := db.Exec(schema)
	return err
}

// storeBootstrap inserts or updates one YAML object in the bootstrap table.
func storeBootstrap(db *sql.DB, id string, node map[string]any, path string) error {
	blob, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	hash := fmt.Sprintf("%x", time.Now().UnixNano()) // placeholder hash until you add proper hashing

	_, err = db.Exec(`
		INSERT INTO bootstrap (id, type, content, source_file, hash)
		VALUES ($1, 'yaml', $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		    SET content = $2,
		        source_file = $3,
		        hash = $4,
		        imported_at = NOW();
	`, id, blob, path, hash)

	return err
}
