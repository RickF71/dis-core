package identity

// import (
// 	"encoding/json"
// 	"net/http"

// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// // HandleIdentityBind handles POST /api/identity/bind requests.
// func HandleIdentityBind(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
// 	ctx := r.Context()

// 	var input struct {
// 		UID    string `json:"uid"`
// 		Domain string `json:"domain"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
// 		http.Error(w, "invalid JSON body", http.StatusBadRequest)
// 		return
// 	}
// 	if input.UID == "" || input.Domain == "" {
// 		http.Error(w, "missing uid or domain", http.StatusBadRequest)
// 		return
// 	}

// 	// Insert or update record in identities table
// 	_, err := db.Exec(ctx, `
// 		INSERT INTO identities (dis_uid, namespace)
// 		VALUES ($1, $2)
// 		ON CONFLICT (dis_uid) DO UPDATE
// 			SET namespace = EXCLUDED.namespace,
// 				updated_at = now();
// 	`, input.UID, input.Domain)
// 	if err != nil {
// 		http.Error(w, "failed to create binding: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]any{
// 		"status":  "success",
// 		"message": "binding recorded",
// 		"uid":     input.UID,
// 		"domain":  input.Domain,
// 	})
// }
