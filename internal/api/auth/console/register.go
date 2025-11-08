package console

import (
	"net/http"

	"github.com/jackc/pgx/v5"
)

// Register wires the console authentication routes to the mux.
func Register(mux *http.ServeMux, store *pgx.Conn) {
	mux.HandleFunc("/api/auth/console/verify", Handle(store))
}
