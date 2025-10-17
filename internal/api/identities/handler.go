package identities

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"dis-core/internal/db"
)

// IdentityPayload represents an incoming or outgoing identity record.
type IdentityPayload struct {
	DISUID    string `json:"dis_uid"`
	Namespace string `json:"namespace"`
}

// writeJSON sends JSON responses with proper headers.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("❌ JSON encode error: %v", err)
	}
}

// HandleIdentities provides both registration (POST) and listing (GET).
func HandleIdentities(store *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		// === POST /api/identities ===
		case http.MethodPost:
			var req IdentityPayload
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if req.DISUID == "" {
				http.Error(w, "Missing dis_uid", http.StatusBadRequest)
				return
			}

			var id int64
			err := store.QueryRow(`
				INSERT INTO identities (dis_uid, namespace, created_at, active)
				VALUES ($1, $2, NOW(), TRUE)
				ON CONFLICT (dis_uid) DO UPDATE
					SET namespace = EXCLUDED.namespace,
					    updated_at = NOW(),
					    active = TRUE
				RETURNING id;
			`, req.DISUID, req.Namespace).Scan(&id)
			if err != nil {
				http.Error(w, fmt.Sprintf("Insert failed: %v", err), http.StatusInternalServerError)
				return
			}

			resp := map[string]any{
				"status":     "created",
				"id":         id,
				"dis_uid":    req.DISUID,
				"namespace":  req.Namespace,
				"created_at": db.NowRFC3339Nano(),
			}
			writeJSON(w, http.StatusCreated, resp)

		// === GET /api/identities?limit=50&offset=0 ===
		case http.MethodGet:
			q := r.URL.Query()
			limit, _ := strconv.Atoi(q.Get("limit"))
			offset, _ := strconv.Atoi(q.Get("offset"))

			if limit <= 0 {
				limit = 50
			}
			if offset < 0 {
				offset = 0
			}

			rows, err := store.Query(`
				SELECT id, dis_uid, namespace, created_at, updated_at, active
				FROM identities
				WHERE active = TRUE
				ORDER BY id DESC
				LIMIT $1 OFFSET $2;
			`, limit, offset)
			if err != nil {
				http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var list []map[string]any
			for rows.Next() {
				var (
					id        int64
					disUID    string
					namespace string
					createdAt string
					updatedAt *string
					active    bool
				)
				if err := rows.Scan(&id, &disUID, &namespace, &createdAt, &updatedAt, &active); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				entry := map[string]any{
					"id":         id,
					"dis_uid":    disUID,
					"namespace":  namespace,
					"created_at": createdAt,
					"updated_at": updatedAt,
					"active":     active,
				}
				list = append(list, entry)
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
