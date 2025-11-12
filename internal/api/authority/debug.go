package authorityapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"dis-core/internal/authority"
)

// HandleListDecisionsDebug is a debug version to check what's in the context
func HandleListDecisionsDebug(w http.ResponseWriter, r *http.Request) {
	// Debug: log all context values
	log.Printf("Context debug: %+v", r.Context())

	// Get database connection from middleware context
	db, ok := GetDBFromContext(r.Context())
	if !ok {
		log.Printf("Failed to get DB from context")
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	log.Printf("Got DB connection: %v", db != nil)

	// Parse query parameters for filtering
	filter := authority.DecisionFilter{}

	if domain := r.URL.Query().Get("domain"); domain != "" {
		filter.Domain = domain
	}

	if policyID := r.URL.Query().Get("policy_id"); policyID != "" {
		filter.PolicyID = policyID
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Set default limit if none specified
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	decisions, err := authority.ListDecisions(r.Context(), db, filter)
	if err != nil {
		log.Printf("ListDecisions error: %v", err)
		http.Error(w, "Failed to list decisions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decisions)
}
