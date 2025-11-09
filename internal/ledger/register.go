package ledger

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Register wires the receipt endpoints into the server mux.
//
// Exposes:
//   - GET /api/receipts  list stored receipts
func Register(mux *http.ServeMux, store *pgxpool.Pool) {
	mux.HandleFunc("/api/receipts", Handle(store))
}
