package auth

import (
	"net/http"

	"dis-core/internal/api/auth/console"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Register wires all authentication routes to the main mux.
func Register(mux *http.ServeMux, store *pgxpool.Pool) {
	console.Register(mux, store)
}
