package ledger

import (
	"fmt"
	"log"
	"strings"
)

// VerifySchema checks whether all expected tables and columns exist in the database.
// It logs discrepancies and returns an error if critical tables are missing.
func (l *Ledger) VerifySchema() error {
	if l.DB == nil {
		return fmt.Errorf("no database connection")
	}

	expected := map[string][]string{
		"receipts":        {"id", "type", "target", "summary", "created_at"},
		"import_receipts": {"id", "type", "target", "summary", "created_at"},
		"domains":         {"id", "name", "parent", "metadata", "created_at"},
		"schemas":         {"id", "name", "version", "definition", "created_at"},
		"overlays":        {"id", "name", "type", "data", "created_at"},
		"policies":        {"id", "name", "rules", "created_at"},
		"mirror_events":   {"id", "message", "hash", "payload", "created_at"},
	}

	log.Println("🔍 Verifying database schema integrity...")
	var problems []string

	for table, cols := range expected {
		var exists bool
		err := l.DB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table).Scan(&exists)
		if err != nil {
			problems = append(problems, fmt.Sprintf("⚠️  failed to check table %s: %v", table, err))
			continue
		}

		if !exists {
			problems = append(problems, fmt.Sprintf("❌ missing table: %s", table))
			continue
		}

		// check columns
		rows, err := l.DB.Query(`
			SELECT column_name
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1`, table)
		if err != nil {
			problems = append(problems, fmt.Sprintf("⚠️  failed to list columns for %s: %v", table, err))
			continue
		}
		defer rows.Close()

		foundCols := []string{}
		for rows.Next() {
			var name string
			_ = rows.Scan(&name)
			foundCols = append(foundCols, name)
		}

		for _, c := range cols {
			if !contains(foundCols, c) {
				problems = append(problems, fmt.Sprintf("❌ table %s missing column: %s", table, c))
			}
		}
	}

	if len(problems) == 0 {
		log.Println("✅ Schema verified: all tables and columns accounted for.")
		return nil
	}

	log.Println("⚠️ Schema verification found issues:")
	for _, p := range problems {
		log.Println("   ", p)
	}

	return fmt.Errorf("schema verification failed: %d issues", len(problems))
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, val) {
			return true
		}
	}
	return false
}
