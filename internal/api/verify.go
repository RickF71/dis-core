package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// HandleVerify processes verification requests from external or internal domains.
// Future versions will validate receipt signatures, trust proofs, or cross-domain attestations.
func HandleVerify(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleVerifyPost(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

func handleVerifyPost(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	log.Printf("🔍 Verify request received: %+v", payload)

	// Placeholder until v0.9.x proof chain logic
	writeJSON(w, http.StatusAccepted, map[string]string{
		"status": "accepted (stubbed)",
		"note":   "verification logic pending in v0.9.x",
	})
}
