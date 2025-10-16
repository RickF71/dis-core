package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"dis-core/internal/db"
)

// HandleIdentityList returns all active identities with optional pagination.
func HandleIdentityList(store *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			limit, _ := strconv.Atoi(q.Get("limit"))
			offset, _ := strconv.Atoi(q.Get("offset"))

			list, err := db.ListIdentities(store, limit, offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp := map[string]any{
				"count": len(list),
				"items": list,
			}
			writeJSON(w, http.StatusOK, resp)

		default:
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		}
	}
}
