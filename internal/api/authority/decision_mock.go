package authorityapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// Decision represents a mock authority decision
type Decision struct {
	ID        string                 `json:"id"`
	Actor     string                 `json:"actor"`
	Domain    string                 `json:"domain"`
	PolicyID  string                 `json:"policy_id"`
	Input     map[string]interface{} `json:"input"`
	Result    map[string]interface{} `json:"result"`
	CreatedAt time.Time              `json:"created_at"`
	Phase     string                 `json:"phase"`
}

// Mock decision data
var mockDecisions = map[string]Decision{
	"dec-001": {
		ID:       "dec-001",
		Actor:    "id-rick",
		Domain:   "domain.user.rick",
		PolicyID: "policy.freeze.control.v1",
		Input: map[string]interface{}{
			"action":   "domain.freeze.v1",
			"resource": "domain.user.rick",
			"context":  map[string]interface{}{"risk_score": 0.2},
		},
		Result: map[string]interface{}{
			"allow":       false,
			"reason":      "deny:freeze:scope",
			"break_glass": false,
		},
		CreatedAt: time.Now().Add(-2 * time.Hour),
		Phase:     "9A-mock",
	},
	"dec-002": {
		ID:       "dec-002",
		Actor:    "id-alice",
		Domain:   "domain.terra",
		PolicyID: "policy.identity.verify.v1",
		Input: map[string]interface{}{
			"action":   "ci.call.v1",
			"resource": "domain.terra",
			"context":  map[string]interface{}{"verification": "pending"},
		},
		Result: map[string]interface{}{
			"allow":       true,
			"reason":      "identity verified",
			"break_glass": false,
		},
		CreatedAt: time.Now().Add(-1 * time.Hour),
		Phase:     "9A-mock",
	},
	"dec-003": {
		ID:       "dec-003",
		Actor:    "system",
		Domain:   "domain.null",
		PolicyID: "policy.risk.assessment.v1",
		Input: map[string]interface{}{
			"action":   "domain.create.v1",
			"resource": "domain.test.new",
			"context":  map[string]interface{}{"risk_score": 0.1},
		},
		Result: map[string]interface{}{
			"allow":       true,
			"reason":      "risk assessment passed",
			"break_glass": false,
		},
		CreatedAt: time.Now().Add(-30 * time.Minute),
		Phase:     "9A-mock",
	},
}

// HandleGetDecisionMock returns mock decision data for GET /api/authority/decision/{id}
func HandleGetDecisionMock(w http.ResponseWriter, r *http.Request) {
	decisionID := chi.URLParam(r, "id")

	if decisionID == "" {
		http.Error(w, "Decision ID required", http.StatusBadRequest)
		return
	}

	decision, exists := mockDecisions[decisionID]
	if !exists {
		http.Error(w, "Decision not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// HandleListEvents returns static list of mock events for GET /api/authority/events
func HandleListEvents(w http.ResponseWriter, r *http.Request) {
	events := []map[string]interface{}{
		{
			"id":         "evt-001",
			"type":       "domain.freeze.v1",
			"domain":     "domain.user.rick",
			"actor":      "id-rick",
			"created_at": time.Now().Add(-3 * time.Hour),
			"phase":      "9A-mock",
		},
		{
			"id":         "evt-002",
			"type":       "ci.call.v1",
			"domain":     "domain.terra",
			"actor":      "id-alice",
			"created_at": time.Now().Add(-2 * time.Hour),
			"phase":      "9A-mock",
		},
		{
			"id":         "evt-003",
			"type":       "domain.unfreeze.v1",
			"domain":     "domain.user.rick",
			"actor":      "system",
			"created_at": time.Now().Add(-1 * time.Hour),
			"phase":      "9A-mock",
		},
		{
			"id":         "evt-004",
			"type":       "policy.evaluate.v1",
			"domain":     "domain.test.new",
			"actor":      "id-bob",
			"created_at": time.Now().Add(-30 * time.Minute),
			"phase":      "9A-mock",
		},
		{
			"id":         "evt-005",
			"type":       "schema.validate.v1",
			"domain":     "domain.null",
			"actor":      "system",
			"created_at": time.Now().Add(-15 * time.Minute),
			"phase":      "9A-mock",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
