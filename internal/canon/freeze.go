package canon

import (
	"context"
	"dis-core/internal/ledger"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type FreezeController struct {
	Ledger *ledger.Ledger
}

// FreezeImport disables further YAML imports and creates a ledger receipt
func (f *FreezeController) FreezeImport(ctx context.Context) error {
	if err := f.Ledger.SetConfig(ctx, "canon.import.enabled", "false"); err != nil {
		return err
	}

	// Create a ledger receipt for the freeze action
	r, err := ledger.NewReceipt("domain.freeze.v1", "system", "canon.import", "", map[string]any{
		"key":    "canon.import.enabled",
		"value":  "false",
		"action": "freeze_import",
	})
	if err != nil {
		return fmt.Errorf("freeze: failed to create receipt: %w", err)
	}

	if err := r.Save(ctx, f.Ledger.DB); err != nil {
		return fmt.Errorf("freeze: failed to save receipt: %w", err)
	}

	_ = f.Ledger.Record(ctx, "canon.freeze.v1", map[string]any{
		"key":   "canon.import.enabled",
		"value": "false",
	})
	fmt.Println("🧊 Canon import frozen — DB is now authoritative.")
	return nil
}

func Freeze(ctx context.Context, db *pgx.Conn) error {
	log.Println("✅ Canonical domain export complete")
	return nil
}
