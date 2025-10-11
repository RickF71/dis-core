package api

import (
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/db"
)

type consoleAuthRequest struct {
	Token     string `json:"token"`
	PolicyRef string `json:"policy_ref,omitempty"` // reserved for future policy binding
}

type consoleAuthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
	// Optionally return a short-lived session or echo subject for UI convenience
	Subject string `json:"subject,omitempty"`
}

func RegisterConsoleAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/console", handleConsoleAuth)
}

// POST /api/auth/console { "token": "<handshake token>" }
func handleConsoleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

	// Optional: rotate, mint short-lived session, etc. For v0.9.3 we just acknowledge.
	resp := consoleAuthResponse{
		Status:  "accepted",
		Time:    now.Format(time.RFC3339),
		Subject: hs.Subject,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
