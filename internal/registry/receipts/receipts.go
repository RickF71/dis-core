package receipts

// stub for receipts registry
import (
	"net/http"

	"github.com/jackc/pgx/v5"
)

func Register(mux *http.ServeMux, db *pgx.Conn) {
	// TODO: implement receipts routes
}
