package atlas

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// Register wires Atlas (geospatial + overlay) routes into the global mux.
func Register(mux *http.ServeMux, store *sql.DB) {
	mux.HandleFunc("/api/atlas/receipts", handleAtlasReceipts)
	mux.HandleFunc("/api/overlay/", handleOverlay)
}

// --- Handlers ---

func handleAtlasReceipts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		resp := map[string]string{
			"status": "ok",
			"note":   "atlas receipts endpoint (v0.9.x stub)",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleOverlay serves overlays for a given domain/scope.
// Later, this will pull merged GeoJSONs from Postgres or cached files.
func handleOverlay(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/api/overlay/"):]
	if path == "" {
		http.Error(w, "missing overlay path", http.StatusBadRequest)
		return
	}

	resp := map[string]string{
		"status":  "ok",
		"overlay": fmt.Sprintf("overlay requested: %s", path),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
