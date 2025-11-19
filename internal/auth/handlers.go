package auth

import (
	"encoding/json"
	"net/http"
)

// WhoAmIResponse represents the sovereign identity summary
// NEVER includes ExternalUID - only bound corporeal domain info
type WhoAmIResponse struct {
	Authenticated      bool   `json:"authenticated"`
	Bound              bool   `json:"bound"`
	CorporealDomainUID string `json:"corporeal_domain_uid,omitempty"`
	Message            string `json:"message,omitempty"`
}

// HandleWhoAmI returns the current user's sovereign identity status
// Endpoint: GET /api/whoami
func HandleWhoAmI(w http.ResponseWriter, r *http.Request) {
	user := GetActiveUser(r)

	var response WhoAmIResponse

	if user == nil || !user.HasExternalUID {
		// No external authentication
		response = WhoAmIResponse{
			Authenticated: false,
			Bound:         false,
			Message:       "No external authentication provided",
		}
	} else if !user.Bound {
		// Authenticated but not bound to corporeal domain
		response = WhoAmIResponse{
			Authenticated: true,
			Bound:         false,
			Message:       "External authentication present but not bound to corporeal domain",
		}
	} else {
		// Authenticated and bound
		response = WhoAmIResponse{
			Authenticated:      true,
			Bound:              true,
			CorporealDomainUID: user.CorporealDomainUID,
			Message:            "Bound to corporeal domain",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ExternalUIDResponse is DEV ONLY - shows the raw external UID for debugging
// WARNING: This endpoint should be removed or disabled in production
type ExternalUIDResponse struct {
	ExternalUID    string `json:"external_uid"`
	HasExternalUID bool   `json:"has_external_uid"`
	Message        string `json:"message"`
}

// HandleWhoAmIExternal returns the raw external UID for development/debugging
// Endpoint: GET /api/whoami/external
//
// DEV ONLY: This endpoint is purely for visual debugging.
// It does NOT persist or log the UID anywhere.
// Should be disabled or removed in production environments.
func HandleWhoAmIExternal(w http.ResponseWriter, r *http.Request) {
	user := GetActiveUser(r)

	var response ExternalUIDResponse

	if user == nil || !user.HasExternalUID {
		response = ExternalUIDResponse{
			ExternalUID:    "",
			HasExternalUID: false,
			Message:        "No external UID present",
		}
	} else {
		response = ExternalUIDResponse{
			ExternalUID:    user.ExternalUID,
			HasExternalUID: true,
			Message:        "External UID from X-External-User header (DEV ONLY - not persisted)",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Dev-Only", "true")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
