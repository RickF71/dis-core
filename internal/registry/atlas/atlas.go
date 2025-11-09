package atlas

// stub for atlas registry
import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(mux *http.ServeMux, db *pgxpool.Pool) {
	// TODO: implement atlas routes
}
