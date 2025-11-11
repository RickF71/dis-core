package identity

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// IdentityBinding represents a simple UID ↔ Domain ↔ Key record
type IdentityBinding struct {
	ID        int       `json:"id"`
	UID       string    `json:"uid"`
	Domain    string    `json:"domain"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}

// HandleCreateIdentityBinding handles POST /api/identity/bind
func HandleCreateIdentityBinding(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	var input struct {
		UID    string `json:"uid"`
		Domain string `json:"domain"`
		Key    string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if input.UID == "" || input.Domain == "" || input.Key == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	var binding IdentityBinding
	query := `
		INSERT INTO identity_bindings (uid, domain, key, created_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (uid, domain) DO UPDATE
			SET key = EXCLUDED.key,
			    created_at = now()
		RETURNING id, uid, domain, key, created_at;
	`

	err := db.QueryRow(ctx, query, input.UID, input.Domain, input.Key).
		Scan(&binding.ID, &binding.UID, &binding.Domain, &binding.Key, &binding.CreatedAt)
	if err != nil {
		http.Error(w, "failed to create binding: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(binding)
}
