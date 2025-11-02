package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// HandleVerify processes verification requests from external or internal domains.
// Future versions will validate receipt signatures, trust proofs, or cross-domain attestations.
func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "verified"})
}

func handleVerifyPost(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	log.Printf("🔍 Verify request received: %+v", payload)

	// Placeholder until v0.9.x proof chain logic
	JSON(w, http.StatusAccepted, map[string]string{
		"status": "accepted (stubbed)",
		"note":   "verification logic pending in v0.9.x",
	})
}
