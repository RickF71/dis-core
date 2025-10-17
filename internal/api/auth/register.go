package auth

import (
	"database/sql"
	"net/http"

	"dis-core/internal/api/auth/console"
	"dis-core/internal/api/auth/dis"
	"dis-core/internal/api/auth/revoke"
)

// Register wires all authentication-related routes.
func Register(mux *http.ServeMux, store *sql.DB) {
	console.Register(mux, store)
	dis.Register(mux, store)
	revoke.Register(mux, store)
}
