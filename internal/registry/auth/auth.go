package auth

// stub for auth registry
import (
	"net/http"

	"github.com/jackc/pgx/v5"
)

func Register(mux *http.ServeMux, db *pgx.Conn) {
	// TODO: implement auth routes
}
