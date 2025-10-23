package canon

import (
	"context"
	"dis-core/internal/config"
	"dis-core/internal/ledger"
)

func Register(ctx context.Context, led *ledger.Ledger, cfg *config.Config) error {
	// TODO: Canon import, freeze, export logic based on flags/config
	return nil
}
