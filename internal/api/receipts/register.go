package receipts

import (
	"database/sql"
	"net/http"
)

// Register wires the receipt endpoints into the server mux.
//
// Exposes:
//   - GET /api/receipts → list stored receipts
func Register(mux *http.ServeMux, store *sql.DB) {
	mux.HandleFunc("/api/receipts", Handle(store))
}
