package identities

import (
	"database/sql"
	"dis-core/internal/db"
	"net/http"
)

// Register wires all identity-related routes to the main mux.
//
// Currently exposes:
//   - POST /api/identities  → create or update an identity
//   - GET  /api/identities  → list active identities
func Register(mux *http.ServeMux, store *sql.DB) {
	mux.HandleFunc("/api/identities", HandleIdentities(db.DefaultDB))
}
