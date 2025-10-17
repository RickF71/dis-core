package api

import (
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/db"
)

// statusPayload represents the JSON structure returned by /api/status.
type statusPayload struct {
	Time   string         `json:"time"`
	Counts map[string]int `json:"counts"`
	Notes  string         `json:"notes,omitempty"`
}

// RegisterStatusRoutes binds /api/status to the global mux.
func RegisterStatusRoutes(mux *http.ServeMux) {
	mux.Handle("/api/status", WithCORS(http.HandlerFunc(HandleStatus)))
}

// HandleStatus responds with a diagnostic summary.
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rcpts, _ := db.CountReceipts()
	hs, _ := db.CountHandshakes()
	rev, _ := db.CountRevocations()
	ids, _ := db.CountIdentities()

	out := statusPayload{
		Time: time.Now().UTC().Format(time.RFC3339),
		Counts: map[string]int{
			"receipts":    int(rcpts),
			"handshakes":  int(hs),
			"revocations": int(rev),
			"identities":  int(ids),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
