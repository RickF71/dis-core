package bootstrap

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ImportAll walks disyaml/ and imports every file into bootstrap_files.
func ImportAll(db *sql.DB, root string) error {
	count := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Skip if somehow we hit outside disyaml (safety belt)
		if !strings.HasPrefix(path, root) {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		filename := filepath.Base(path)

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}

		_, err = db.Exec(`
			INSERT INTO bootstrap_files (rel_path, filename, content)
			VALUES ($1, $2, $3)
		`, rel, filename, data)
		if err != nil {
			return fmt.Errorf("insert %s: %w", rel, err)
		}

		count++
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("✅ Imported %d files from %s\n", count, root)
	return nil
}
