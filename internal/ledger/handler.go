package ledger

import (
	"encoding/json"
	"net/http"
	"strconv"

	"dis-core/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Handle returns an http.HandlerFunc bound to the provided pgxpool connection.
func Handle(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			limit, _ := strconv.Atoi(q.Get("limit"))
			offset, _ := strconv.Atoi(q.Get("offset"))
			if limit <= 0 {
				limit = 100
			}

			ctx := r.Context()
			list, err := db.ListReceipts(ctx, pool, limit, offset)
			if err != nil {
				http.Error(w, "failed to list receipts: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count": len(list),
				"items": list,
			})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
