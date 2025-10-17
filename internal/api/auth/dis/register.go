package dis

import (
	"database/sql"
	"net/http"
)

// Register wires DIS handshake endpoints to the mux.
//
// Exposes:
//   - POST /api/auth/dis → create handshake
//   - GET  /api/auth/dis → list handshakes
func Register(mux *http.ServeMux, store *sql.DB) {
	mux.HandleFunc("/api/auth/dis", Handle(store))
}
