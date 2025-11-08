package auth

import (
	"net/http"

	"github.com/jackc/pgx/v5"

	"dis-core/internal/api/auth/console"
	"dis-core/internal/api/auth/dis"
	"dis-core/internal/api/auth/revoke"
)

// Register wires all authentication-related routes.
func Register(mux *http.ServeMux, store *pgx.Conn) {
	console.Register(mux, store)
	dis.Register(mux, store)
	revoke.Register(mux, store)
}
