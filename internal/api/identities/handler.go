package identities

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Identity represents a minimal identity record for DIS-Core.
type Identity struct {
	ID        string            `json:"id"`
	Claims    map[string]string `json:"claims"`
	CreatedAt time.Time         `json:"created_at"`
}

// HandleIdentities provides REST operations for identities.
// GET    → lists all identities
// POST   → creates or updates an identity (upsert)
func HandleIdentities(store *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListIdentities(w, r, store)
		case http.MethodPost:
			handleCreateIdentity(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleListIdentities retrieves all identities from the database.
func handleListIdentities(w http.ResponseWriter, r *http.Request, store *pgxpool.Pool) {
	ctx := r.Context()

	rows, err := store.Query(ctx, `
		SELECT id, claims, created_at
		FROM identities
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("failed to query identities: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var identities []Identity
	for rows.Next() {
		var identity Identity
		var claimsJSON string

		if err := rows.Scan(&identity.ID, &claimsJSON, &identity.CreatedAt); err != nil {
			log.Printf("failed to scan identity: %v", err)
			continue
		}

		if err := json.Unmarshal([]byte(claimsJSON), &identity.Claims); err != nil {
			log.Printf("failed to unmarshal claims: %v", err)
			identity.Claims = make(map[string]string)
		}

		identities = append(identities, identity)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(identities)
}

// handleCreateIdentity creates or updates an identity.
func handleCreateIdentity(w http.ResponseWriter, r *http.Request, store *pgxpool.Pool) {
	ctx := r.Context()

	var identity Identity
	if err := json.NewDecoder(r.Body).Decode(&identity); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if identity.ID == "" {
		http.Error(w, "identity ID is required", http.StatusBadRequest)
		return
	}

	claimsJSON, err := json.Marshal(identity.Claims)
	if err != nil {
		http.Error(w, "failed to marshal claims", http.StatusInternalServerError)
		return
	}

	// Upsert the identity
	_, err = store.Exec(ctx, `
		INSERT INTO identities (id, claims, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (id)
		DO UPDATE SET claims = EXCLUDED.claims
	`, identity.ID, string(claimsJSON))

	if err != nil {
		log.Printf("failed to upsert identity: %v", err)
		http.Error(w, "failed to save identity", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"status": "created", "id": "%s"}`, identity.ID)
}
