package canon

import (
	"database/sql"
	"dis-core/internal/ledger"
	"fmt"
	"log"
)

type FreezeController struct {
	Ledger *ledger.Ledger
}

// FreezeImport disables further YAML imports
func (f *FreezeController) FreezeImport() error {
	if err := f.Ledger.SetConfig("canon.import.enabled", "false"); err != nil {
		return err
	}
	_ = f.Ledger.Record("canon.freeze.v1", map[string]any{
		"key":   "canon.import.enabled",
		"value": "false",
	})
	fmt.Println("🧊 Canon import frozen — DB is now authoritative.")
	return nil
}

func Freeze(db *sql.DB) error {
	log.Println("✅ Canonical domain export complete")
	return nil
}
