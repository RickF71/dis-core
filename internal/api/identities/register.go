package identities

import (
	"net/http"

	"github.com/jackc/pgx/v5"
)

// Register wires all identity-related routes to the main mux.
//
// Currently exposes:
//   - POST /api/identities  → create or update an identity
//   - GET  /api/identities  → list active identities
func Register(mux *http.ServeMux, store *pgx.Conn) {
	mux.HandleFunc("/api/identities", HandleIdentities(store))
}
