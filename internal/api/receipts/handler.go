package ledger

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"dis-core/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Handle returns an http.HandlerFunc bound to the provided DB connection.
func Handle(store *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ctx := context.Background()
			q := r.URL.Query()
			limit, _ := strconv.Atoi(q.Get("limit"))
			offset, _ := strconv.Atoi(q.Get("offset"))
			if limit <= 0 {
				limit = 100
			}

			list, err := db.ListReceipts(ctx, store, limit, offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
