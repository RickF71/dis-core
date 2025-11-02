package importer

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ExportCanon writes canon entries back into domain folders under /canon/
func ExportCanon(root string, db *sql.DB) error {
	canonDir := filepath.Join(root, "canon")
	err := os.MkdirAll(canonDir, 0755)
	if err != nil {
		return err
	}

	rows, err := db.Query(`SELECT type, version, domain, content FROM canon`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var typ, version, domain string
		var content map[string]any
		if err := rows.Scan(&typ, &version, &domain, &content); err != nil {
			return err
		}

		outDir := filepath.Join(canonDir, domain)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return err
		}

		outPath := filepath.Join(outDir, fmt.Sprintf("%s.yaml", typ))
		data, _ := yaml.Marshal(content)
		err = os.WriteFile(outPath, data, 0644)
		if err != nil {
			return fmt.Errorf("write canon file: %w", err)
		}
		fmt.Printf("💾 Exported canon: %s\n", outPath)
	}
	return nil
}
