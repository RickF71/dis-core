package canon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"

	"dis-core/internal/ledger"
)

// CanonImporter handles YAML → DB import
type CanonImporter struct {
	Ledger *ledger.Ledger
}

// CanonRecord represents a parsed YAML object
type CanonRecord struct {
	ID      string         `yaml:"id"`
	Type    string         `yaml:"type"`
	Version string         `yaml:"version"`
	Content map[string]any `yaml:",inline"`
	Meta    map[string]any `yaml:"_meta"`
	Hash    string         `yaml:"-"`
}

// ImportDir walks a directory and imports each YAML into the ledger
func (c *CanonImporter) ImportDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("📂 No YAML files found in", dir)
		return nil
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}

		h := sha256.Sum256(data)
		record := CanonRecord{}
		if err := yaml.Unmarshal(data, &record); err != nil {
			_ = c.Ledger.RecordCall(context.Background(), "system", f, "canon", "import.failed.v1", map[string]any{
				"file":  f,
				"error": err.Error(),
				"hash":  hex.EncodeToString(h[:]),
			})
			continue
		}

		record.Hash = hex.EncodeToString(h[:])
		record.Meta = map[string]any{
			"source_file": filepath.Base(f),
			"imported_at": time.Now().UTC().Format(time.RFC3339),
			"hash":        record.Hash,
		}

		if err := c.Ledger.StoreCanon(context.Background(), record); err != nil {
			_ = c.Ledger.RecordCall(context.Background(), "system", f, "canon", "store.failed.v1", map[string]any{
				"file":  f,
				"error": err.Error(),
				"hash":  record.Hash,
			})
			continue
		}

		_ = c.Ledger.RecordCall(context.Background(), "system", f, "canon", "import.success.v1", map[string]any{
			"file": f,
			"hash": record.Hash,
			"type": record.Type,
			"id":   record.ID,
		})
		fmt.Printf("✅ Imported %s (%s)\n", f, record.ID)
	}
	return nil
}

// Import initializes CanonImporter with the given pgxpool connection
func Import(pool *pgxpool.Pool) error {
	log.Println("📜 Canon import starting")
	led := &ledger.Ledger{DB: pool}
	importer := &CanonImporter{Ledger: led}
	// In the future, pass directory path dynamically
	err := importer.ImportDir("./canon")
	if err != nil {
		return fmt.Errorf("canon import failed: %w", err)
	}
	log.Println("📜 Canon import complete")
	return nil
}
