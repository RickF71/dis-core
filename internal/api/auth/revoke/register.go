package revoke

import (
	"database/sql"
	"net/http"
)

// Register wires revocation routes to the mux.
func Register(mux *http.ServeMux, store *sql.DB) {
	mux.HandleFunc("/api/auth/revoke", Handle(store))
}
