package dis

import (
	"net/http"

	"github.com/jackc/pgx/v5"
)

// Register wires DIS handshake endpoints to the mux.
//
// Exposes:
//   - POST /api/auth/dis → create handshake
//   - GET  /api/auth/dis → list handshakes
func Register(mux *http.ServeMux, store *pgx.Conn) {
	mux.HandleFunc("/api/auth/dis", Handle(store))
}
