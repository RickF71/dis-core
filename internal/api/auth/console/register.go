package console

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Register wires the console authentication routes to the mux.
func Register(mux *http.ServeMux, store *pgxpool.Pool) {
	mux.HandleFunc("/api/auth/console/verify", Handle(store))
}
