package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type ConsoleAuthSession struct {
	SessionID  string   `json:"session_id"`
	Initiator  string   `json:"initiator_dis"`
	Privileges []string `json:"privileges"`
	Active     bool     `json:"active"`
	ExpiresAt  string   `json:"expires_at"`
}

func HandleConsoleAuth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var sess ConsoleAuthSession
		if err := json.NewDecoder(r.Body).Decode(&sess); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sess.SessionID = "sess-" + time.Now().Format("20060102150405")
		sess.Active = true
		sess.ExpiresAt = time.Now().Add(2 * time.Hour).Format(time.RFC3339)
		json.NewEncoder(w).Encode(sess)
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}
