package api

import (
	"encoding/json"
	"net/http"
)

// HandleExternalVerify is a placeholder for external verification endpoints.
// In v0.8.1+, this will verify inbound receipts or policy calls from other domains.
func HandleExternalVerify(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleExternalVerifyPost(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func handleExternalVerifyPost(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	// Placeholder: simulate external verification
	writeJSON(w, http.StatusAccepted, map[string]string{
		"status": "accepted (stubbed)",
		"note":   "external verification logic pending in v0.8.1",
	})
}
