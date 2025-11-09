package receipts

// stub for receipts registry
import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(mux *http.ServeMux, db *pgxpool.Pool) {
	// TODO: implement receipts routes
}
