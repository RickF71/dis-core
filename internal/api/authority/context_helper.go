package authorityapi

import (
	"context"

	"dis-core/internal/api/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetDBFromContext retrieves the database connection from middleware context
func GetDBFromContext(ctx context.Context) (*pgxpool.Pool, bool) {
	// Use the middleware's FromDB function
	if db := middleware.FromDB(ctx); db != nil {
		return db, true
	}

	return nil, false
}
