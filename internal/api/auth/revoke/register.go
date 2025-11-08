package revoke

import (
	"net/http"

	"github.com/jackc/pgx/v5"
)

// Register wires revocation routes to the mux.
func Register(mux *http.ServeMux, store *pgx.Conn) {
	mux.HandleFunc("/api/auth/revoke", Handle(store))
}
