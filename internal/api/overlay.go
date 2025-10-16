// internal/api/overlay.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// OverlayData represents the schema envelope for overlay responses.
type OverlayData struct {
	Schema  string      `json:"$schema"`
	Domain  string      `json:"domain"`
	Scope   string      `json:"scope"`
	Version string      `json:"version"`
	Data    interface{} `json:"data"`
}

// GetOverlayHandler serves overlay data for a given domain and scope.
// Example: /api/overlay/domain.terra/authority
func GetOverlayHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/overlay/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid overlay path", http.StatusBadRequest)
		return
	}

	domain := parts[0]
	scope := parts[1]

	// Retrieve overlay data from ledger (mocked for now)
	data, err := BuildOverlay(domain, scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := OverlayData{
		Schema:  "overlay.v0.1",
		Domain:  domain,
		Scope:   scope,
		Version: "v0.9.6",
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(resp)
}

// BuildOverlay provides mock overlay data for the requested domain and scope.
// This will later be replaced by real ledger logic.
func BuildOverlay(domain, scope string) (interface{}, error) {
	// Example mock dataset — replace with real query when ledger integration is ready
	mock := map[string]interface{}{
		"summary": fmt.Sprintf("Overlay for %s/%s", domain, scope),
		"nodes": []map[string]string{
			{"id": "seat.root", "label": "Root Seat"},
			{"id": "seat.gov", "label": "Government Seat"},
			{"id": "seat.user", "label": "Citizen Seat"},
		},
		"edges": []map[string]string{
			{"from": "seat.root", "to": "seat.gov", "type": "delegation"},
			{"from": "seat.gov", "to": "seat.user", "type": "representation"},
		},
	}

	// Optional: use scope to tailor the overlay structure
	switch scope {
	case "authority":
		mock["overlayType"] = "authority-chain"
	case "trust":
		mock["overlayType"] = "trust-graph"
	case "consent":
		mock["overlayType"] = "consent-gate"
	case "freeze":
		mock["overlayType"] = "freeze-state"
	default:
		mock["overlayType"] = "generic"
	}

	return mock, nil
}
