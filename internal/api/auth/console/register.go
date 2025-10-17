package console

import (
	"database/sql"
	"net/http"
)

// Register wires the console authentication routes to the mux.
func Register(mux *http.ServeMux, store *sql.DB) {
	mux.HandleFunc("/api/auth/console/verify", Handle(store))
}
