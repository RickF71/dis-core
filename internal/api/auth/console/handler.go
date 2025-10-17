package console

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/db"
)

// consoleAuthRequest is sent by clients to verify a session or handshake token.
type consoleAuthRequest struct {
	Token     string `json:"token"`
	PolicyRef string `json:"policy_ref,omitempty"`
}

type consoleAuthResponse struct {
	Status  string `json:"status"`
	Time    string `json:"time"`
	Subject string `json:"subject,omitempty"`
}

// Handle returns an http.HandlerFunc that verifies a console token
// and returns acceptance or rejection.
func Handle(store *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req consoleAuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		hs, err := db.GetHandshakeByToken(req.Token)
		if err != nil || hs.ID == 0 {
			http.Error(w, "unknown token", http.StatusUnauthorized)
			return
		}

		now := time.Now().UTC()
		if !hs.ExpiresAt.IsZero() && now.After(hs.ExpiresAt) {
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}
		if !hs.RevokedAt.IsZero() {
			http.Error(w, "token revoked", http.StatusUnauthorized)
			return
		}

		resp := consoleAuthResponse{
			Status:  "accepted",
			Time:    now.Format(time.RFC3339),
			Subject: hs.Subject,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
