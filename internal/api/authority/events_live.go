package authorityapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Event represents a live authority event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Domain    string                 `json:"domain"`
	Actor     string                 `json:"actor"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	Phase     string                 `json:"phase"`
}

// HandleListEventsLive returns live event data from the database
func HandleListEventsLive(w http.ResponseWriter, r *http.Request) {
	// Get database connection from middleware context
	db, ok := GetDBFromContext(r.Context())
	if !ok {
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	// Query recent authority decisions as events
	query := `
		SELECT id, actor, domain, policy_id, input, result, created_at, phase_tag
		FROM authority_decisions
		ORDER BY created_at DESC
		LIMIT 20
	`

	rows, err := db.Query(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to query events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var decision_id, actor, domain, policy_id, phase_tag string
		var input_json, result_json []byte
		var created_at time.Time

		err := rows.Scan(&decision_id, &actor, &domain, &policy_id, &input_json, &result_json, &created_at, &phase_tag)
		if err != nil {
			continue // Skip rows with errors
		}

		// Parse input and result
		var input, result map[string]interface{}
		json.Unmarshal(input_json, &input)
		json.Unmarshal(result_json, &result)

		// Create event from decision
		event := Event{
			ID:        fmt.Sprintf("evt-%s", decision_id[:8]), // Shortened ID for events
			Type:      "authority.decision.v1",
			Domain:    domain,
			Actor:     actor,
			CreatedAt: created_at,
			Phase:     phase_tag,
			Payload: map[string]interface{}{
				"policy_id": policy_id,
				"input":     input,
				"result":    result,
			},
		}

		events = append(events, event)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// queryRecentEvents fetches recent authority events from various tables
func queryRecentEvents(ctx context.Context, db *pgxpool.Pool) ([]Event, error) {
	var allEvents []Event

	// Get decision events
	decisionEvents, err := getDecisionEvents(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision events: %w", err)
	}
	allEvents = append(allEvents, decisionEvents...)

	// TODO: Add other event sources (mirror_events, import events, etc.)

	return allEvents, nil
}

// getDecisionEvents converts authority decisions to events
func getDecisionEvents(ctx context.Context, db *pgxpool.Pool) ([]Event, error) {
	query := `
		SELECT id, actor, domain, policy_id, created_at, phase_tag
		FROM authority_decisions
		ORDER BY created_at DESC
		LIMIT 10
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var id, actor, domain, policy_id, phase_tag string
		var created_at time.Time

		err := rows.Scan(&id, &actor, &domain, &policy_id, &created_at, &phase_tag)
		if err != nil {
			continue
		}

		event := Event{
			ID:        fmt.Sprintf("evt-dec-%s", id[:8]),
			Type:      "authority.decision.v1",
			Domain:    domain,
			Actor:     actor,
			CreatedAt: created_at,
			Phase:     phase_tag,
			Payload: map[string]interface{}{
				"policy_id":   policy_id,
				"decision_id": id,
			},
		}

		events = append(events, event)
	}

	return events, nil
}
