package receipts

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"dis-core/internal/db"
)

// Handle returns an http.HandlerFunc bound to the provided DB connection.
func Handle(store *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			limit, _ := strconv.Atoi(q.Get("limit"))
			offset, _ := strconv.Atoi(q.Get("offset"))
			if limit <= 0 {
				limit = 100
			}

			list, err := db.ListReceipts(store, db.ListOpts{
				Limit:  limit,
				Offset: offset,
			})
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
