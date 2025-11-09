package identities

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Register wires all identity-related routes to the main mux.
//
// Currently exposes:
//   - POST /api/identities  → create or update an identity
//   - GET  /api/identities  → list active identities
func Register(mux *http.ServeMux, store *pgxpool.Pool) {
	mux.HandleFunc("/api/identities", HandleIdentities(store))
}
