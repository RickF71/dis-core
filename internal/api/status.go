package api

import (
	"dis-core/internal/db"
	"encoding/json"
	"net/http"
	"time"
)

type statusPayload struct {
	Time   string         `json:"time"`
	Counts map[string]int `json:"counts"`
	Notes  string         `json:"notes,omitempty"`
}

// RegisterStatusRoutes adds /api/status.
func RegisterStatusRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/status", HandleStatus)
}

func HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	now := time.Now().UTC()

	rcpts, _ := db.CountReceipts()
	hs, _ := db.CountHandshakes()
	rev, _ := db.CountRevocations() // or derive from handshakes where revoked_at not null
	ids, _ := db.CountIdentities()

	out := statusPayload{
		Time: now.Format(time.RFC3339),
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
