package bridge

import (
	"encoding/json"
	"net/http"
)

// GetOverlayHandler exposes the Terra overlay endpoint without creating circular imports.
// Other modules can call bridge.GetOverlayHandler directly.
func GetOverlayHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"status":  "ok",
		"message": "Overlay handler active (bridge module stub)",
		"path":    r.URL.Path,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
