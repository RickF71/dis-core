package authorityapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"dis-core/internal/authority"
)

// HandleGetDecision returns live decision data for GET /api/authority/decision/{id}
func HandleGetDecision(w http.ResponseWriter, r *http.Request) {
	decisionID := chi.URLParam(r, "id")

	if decisionID == "" {
		http.Error(w, "Decision ID required", http.StatusBadRequest)
		return
	}

	// Get database connection from middleware context
	db, ok := GetDBFromContext(r.Context())
	if !ok {
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	decision, err := authority.GetDecision(r.Context(), db, decisionID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			http.Error(w, "Decision not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve decision", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// HandleListDecisions returns live decision list for GET /api/authority/decisions
func HandleListDecisions(w http.ResponseWriter, r *http.Request) {
	// Get database connection from middleware context
	db, ok := GetDBFromContext(r.Context())
	if !ok {
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

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
		http.Error(w, "Failed to list decisions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decisions)
}

// HandlePostDecision creates a new decision for POST /api/authority/decision
func HandlePostDecision(w http.ResponseWriter, r *http.Request) {
	// Get database connection from middleware context
	db, ok := GetDBFromContext(r.Context())
	if !ok {
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	// Get actor from context (set by auth middleware)
	actor, ok := r.Context().Value("actor").(string)
	if !ok {
		actor = "unknown" // Fallback for now
	}

	// Parse request body
	var req struct {
		Domain   string                 `json:"domain"`
		PolicyID string                 `json:"policy_id"`
		Input    map[string]interface{} `json:"input"`
		Result   map[string]interface{} `json:"result"`
		Reason   string                 `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Domain == "" {
		http.Error(w, "Domain is required", http.StatusBadRequest)
		return
	}
	if req.PolicyID == "" {
		http.Error(w, "PolicyID is required", http.StatusBadRequest)
		return
	}
	if req.Input == nil {
		http.Error(w, "Input is required", http.StatusBadRequest)
		return
	}
	if req.Result == nil {
		http.Error(w, "Result is required", http.StatusBadRequest)
		return
	}

	// Create decision
	decision := &authority.Decision{
		Actor:    actor,
		Domain:   req.Domain,
		PolicyID: req.PolicyID,
		Input:    req.Input,
		Result:   req.Result,
		Reason:   req.Reason,
		PhaseTag: "9B-live",
	}

	if err := authority.StoreDecision(r.Context(), db, decision); err != nil {
		http.Error(w, "Failed to store decision", http.StatusInternalServerError)
		return
	}

	// Create provenance receipt
	provenance := ProvenanceInfo{
		RequestID:    r.Header.Get("X-Request-ID"),
		UserAgent:    r.Header.Get("User-Agent"),
		RemoteAddr:   r.RemoteAddr,
		Timestamp:    time.Now(),
		PolicyEngine: "opa",      // TODO: Get from config
		SourceChain:  []string{}, // TODO: Extract from request context
	}

	err := CreateProvenanceReceipt(r.Context(), db, decision, provenance)
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to create provenance receipt: %v", err)
	}

	// Broadcast live event via WebSocket
	eventPayload := map[string]interface{}{
		"decision_id": decision.ID,
		"domain":      decision.Domain,
		"actor":       decision.Actor,
		"policy_id":   decision.PolicyID,
		"result":      decision.Result,
	}
	BroadcastDecisionEvent(eventPayload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(decision)
}
