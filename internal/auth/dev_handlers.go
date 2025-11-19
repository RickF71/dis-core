package auth

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─────────────────────────────────────────────────────────────
// DEV USER STRUCTS
// ─────────────────────────────────────────────────────────────

type DevUser struct {
	ID   string `json:"id"` // MUST be UUID
	Name string `json:"name"`
}

type DevUsersResponse struct {
	Users []DevUser `json:"users"`
}

type DevLoginRequest struct {
	UserID string `json:"user_id"`
}

type DevLoginResponse struct {
	OK      bool   `json:"ok"`
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

// ─────────────────────────────────────────────────────────────
// GET /api/auth/dev-users
// Returns *real UUID identities* from DB
// ─────────────────────────────────────────────────────────────

func HandleDevUsers(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		rows, err := db.Query(ctx,
			`SELECT id, display_name
             FROM identities
             ORDER BY display_name ASC
             LIMIT 50`)
		if err != nil {
			http.Error(w, "Failed to query identities", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		users := []DevUser{}
		for rows.Next() {
			var id uuid.UUID
			var name string

			if err := rows.Scan(&id, &name); err != nil {
				http.Error(w, "Failed to scan identity row", http.StatusInternalServerError)
				return
			}

			users = append(users, DevUser{
				ID:   id.String(),
				Name: name,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DevUsersResponse{Users: users})
	}
}

// ─────────────────────────────────────────────────────────────
// POST /api/auth/dev-login
// Validates that the UUID is a real identity
// ─────────────────────────────────────────────────────────────

func HandleDevLogin(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DevLoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate that this UUID exists in identities
		var exists bool
		err := db.QueryRow(
			r.Context(),
			"SELECT EXISTS(SELECT 1 FROM identities WHERE id=$1)",
			req.UserID,
		).Scan(&exists)

		if err != nil || !exists {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}

		// Return success — frontend will set X-External-User
		resp := DevLoginResponse{
			OK:      true,
			UserID:  req.UserID,
			Message: "Dev login successful",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
