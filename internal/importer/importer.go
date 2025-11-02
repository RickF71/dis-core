package importer

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Extracts version number from filename (e.g., .v1.0.yaml)
func extractVersionFromFilename(path string) string {
	base := filepath.Base(path)
	re := regexp.MustCompile(`\.v(\d+\.\d+)\.ya?ml$`)
	match := re.FindStringSubmatch(base)
	if len(match) > 1 {
		return "v" + match[1]
	}
	return ""
}

// Detects whether a YAML object is a schema definition or domain definition
func detectFileKind(obj map[string]any) string {
	t := fmt.Sprint(obj["type"])
	switch {
	case strings.HasPrefix(t, "domain.") && strings.Contains(t, "schema"):
		return "schema"
	case strings.HasPrefix(t, "domain."):
		return "domain"
	default:
		return "unknown"
	}
}

// ImportBootstrap walks the bootstrap directory, validates and stores canon
func ImportBootstrap(root string, db *sql.DB) error {
	bootstrapDir := filepath.Join(root, "bootstrap")
	archiveDir := filepath.Join(root, "disyamlimp")

	err := os.MkdirAll(archiveDir, 0755)
	if err != nil {
		return fmt.Errorf("create archive dir: %w", err)
	}

	return filepath.WalkDir(bootstrapDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".json") {
			return nil
		}

		fmt.Printf("📥 Importing bootstrap file: %s\n", path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var obj map[string]any
		if strings.HasSuffix(path, ".yaml") {
			err = yaml.Unmarshal(data, &obj)
		} else {
			err = json.Unmarshal(data, &obj)
		}
		if err != nil {
			return fmt.Errorf("parse file %s: %w", path, err)
		}

		fileKind := detectFileKind(obj)
		hash := fmt.Sprintf("%x", sha256.Sum256(data))

		switch fileKind {
		case "schema":
			// Extract key fields
			typ := fmt.Sprint(obj["type"])
			domain := fmt.Sprint(obj["domain"])
			canon := false
			if c, ok := obj["canon"].(bool); ok {
				canon = c
			}

			version := fmt.Sprint(obj["version"])
			fileVersion := extractVersionFromFilename(path)
			if fileVersion != "" {
				version = fileVersion
				fmt.Printf("📦 Using version %s from filename for %s\n", version, path)
			}

			if version == "" {
				return fmt.Errorf("❌ missing version for schema: %s", path)
			}
			if domain == "" {
				return fmt.Errorf("❌ missing domain for schema: %s", path)
			}

			id := fmt.Sprintf("%s@%s", typ, domain)
			_, err = db.Exec(`
				INSERT INTO schemas (id, type, version, domain, source_file, hash, content, imported_at, canon)
				VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), $8)
				ON CONFLICT (id) DO NOTHING;
			`, id, typ, version, domain, path, hash, obj, canon)
			if err != nil {
				return fmt.Errorf("db insert failed: %w", err)
			}

			fmt.Printf("✅ Imported schema: %s (%s)\n", typ, version)

		case "domain":
			// Versionless import: domains are structural, not versioned
			id := fmt.Sprint(obj["domain"])
			parent := fmt.Sprint(obj["parent_domain"])

			if id == "" {
				return fmt.Errorf("❌ missing domain id in %s", path)
			}

			_, err = db.Exec(`
				INSERT INTO domains (id, parent_id, in_tree, canon, imported_at)
				VALUES ($1, $2, FALSE, FALSE, NOW())
				ON CONFLICT (id) DO NOTHING;
			`, id, parent)
			if err != nil {
				return fmt.Errorf("db insert failed for domain %s: %w", id, err)
			}
			fmt.Printf("🏛 Imported domain: %s (parent: %s)\n", id, parent)

		default:
			fmt.Printf("⚠️  Unknown file type: %s\n", path)
		}

		// Move to archive after import
		dest := filepath.Join(archiveDir, filepath.Base(path))
		err = os.Rename(path, dest)
		if err != nil {
			fmt.Printf("⚠️  Could not move file to archive: %v\n", err)
		} else {
			fmt.Printf("📦 Archived %s → %s\n", filepath.Base(path), dest)
		}
		return nil
	})
}
