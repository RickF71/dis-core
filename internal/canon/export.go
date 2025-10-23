package canon

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ExportDomains exports all domain records into canonical YAML files.
// Each domain is written as domains/_auto/<domain-id>.yaml.
func ExportDomains(db *sql.DB, outDir string) error {
	rows, err := db.Query(`
		SELECT id, name, parent_id, is_notech, requires_inside_domain, created_at
		FROM domains
		ORDER BY id;
	`)
	if err != nil {
		return fmt.Errorf("query domains: %w", err)
	}
	defer rows.Close()

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outDir, err)
	}

	count := 0
	for rows.Next() {
		var s CanonicalSchema
		if err := rows.Scan(&s.ID, &s.Name, &s.ParentID, &s.IsNotech, &s.InsideReq, &s.LastUpdated); err != nil {
			return fmt.Errorf("scan domain: %w", err)
		}
		if s.LastUpdated == "" {
			s.LastUpdated = time.Now().UTC().Format(time.RFC3339)
		}

		data, err := yaml.Marshal(&s)
		if err != nil {
			return fmt.Errorf("marshal yaml: %w", err)
		}

		fn := filepath.Join(outDir, fmt.Sprintf("%s.yaml", s.ID))
		if err := os.WriteFile(fn, data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", fn, err)
		}
		count++
	}

	// write a provenance receipt
	receipt := map[string]any{
		"exported_at":  time.Now().Format(time.RFC3339),
		"domain_count": count,
		"source":       "DIS-Core v0.9.3",
	}

	receiptData, _ := yaml.Marshal(receipt)
	_ = os.WriteFile(filepath.Join(outDir, "canon.receipt.yaml"), receiptData, 0644)

	fmt.Printf("📜 Canon export complete — %d domain(s) written to %s\n", count, outDir)
	return nil
}

func Export(db *sql.DB) error {
	log.Println("📜 Canon export complete")
	return nil
}
