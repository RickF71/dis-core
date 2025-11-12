package authority

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Decision represents an authority decision stored in the database
type Decision struct {
	ID         string                 `json:"id" db:"id"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	Actor      string                 `json:"actor" db:"actor"`
	Domain     string                 `json:"domain" db:"domain"`
	PolicyID   string                 `json:"policy_id" db:"policy_id"`
	Input      map[string]interface{} `json:"input" db:"input"`
	Result     map[string]interface{} `json:"result" db:"result"`
	Reason     string                 `json:"reason" db:"reason"`
	PhaseTag   string                 `json:"phase_tag" db:"phase_tag"`
	ReplayHash string                 `json:"replay_hash" db:"replay_hash"`
}

// DecisionFilter provides filtering options for decision queries
type DecisionFilter struct {
	Domain   string `json:"domain,omitempty"`
	PolicyID string `json:"policy_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// StoreDecision stores a new decision in the database
func StoreDecision(ctx context.Context, db *pgxpool.Pool, decision *Decision) error {
	// Generate ID if empty
	if decision.ID == "" {
		decision.ID = uuid.New().String()
	}

	// Set creation time if not set
	if decision.CreatedAt.IsZero() {
		decision.CreatedAt = time.Now()
	}

	// Compute replay hash
	decision.ReplayHash = computeReplayHash(decision.Input, decision.Result, decision.PolicyID)

	// Convert maps to JSON for storage
	inputJSON, err := json.Marshal(decision.Input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	resultJSON, err := json.Marshal(decision.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	query := `
		INSERT INTO authority_decisions (
			id, created_at, actor, domain, policy_id,
			input, result, reason, phase_tag, replay_hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = db.Exec(ctx, query,
		decision.ID,
		decision.CreatedAt,
		decision.Actor,
		decision.Domain,
		decision.PolicyID,
		inputJSON,
		resultJSON,
		decision.Reason,
		decision.PhaseTag,
		decision.ReplayHash,
	)

	return err
}

// GetDecision retrieves a decision by ID from the database
func GetDecision(ctx context.Context, db *pgxpool.Pool, id string) (*Decision, error) {
	query := `
		SELECT id, created_at, actor, domain, policy_id,
			   input, result, reason, phase_tag, replay_hash
		FROM authority_decisions
		WHERE id = $1
	`

	row := db.QueryRow(ctx, query, id)

	var decision Decision
	var inputJSON, resultJSON []byte

	err := row.Scan(
		&decision.ID,
		&decision.CreatedAt,
		&decision.Actor,
		&decision.Domain,
		&decision.PolicyID,
		&inputJSON,
		&resultJSON,
		&decision.Reason,
		&decision.PhaseTag,
		&decision.ReplayHash,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(inputJSON, &decision.Input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &decision.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &decision, nil
}

// ListDecisions retrieves decisions based on filter criteria
func ListDecisions(ctx context.Context, db *pgxpool.Pool, filter DecisionFilter) ([]*Decision, error) {
	query := `
		SELECT id, created_at, actor, domain, policy_id,
			   input, result, reason, phase_tag, replay_hash
		FROM authority_decisions
	`

	args := []interface{}{}
	argCount := 0
	whereClauses := []string{}

	if filter.Domain != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("domain = $%d", argCount))
		args = append(args, filter.Domain)
	}

	if filter.PolicyID != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("policy_id = $%d", argCount))
		args = append(args, filter.PolicyID)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for _, clause := range whereClauses[1:] {
			query += " AND " + clause
		}
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []*Decision
	for rows.Next() {
		var decision Decision
		var inputJSON, resultJSON []byte

		err := rows.Scan(
			&decision.ID,
			&decision.CreatedAt,
			&decision.Actor,
			&decision.Domain,
			&decision.PolicyID,
			&inputJSON,
			&resultJSON,
			&decision.Reason,
			&decision.PhaseTag,
			&decision.ReplayHash,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(inputJSON, &decision.Input); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input for decision %s: %w", decision.ID, err)
		}

		if err := json.Unmarshal(resultJSON, &decision.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result for decision %s: %w", decision.ID, err)
		}

		decisions = append(decisions, &decision)
	}

	return decisions, rows.Err()
}

// computeReplayHash computes SHA256 hash of input+result+policy_id for replay verification
func computeReplayHash(input, result map[string]interface{}, policyID string) string {
	hasher := sha256.New()

	// Marshal input and result to get consistent JSON representation
	inputJSON, _ := json.Marshal(input)
	resultJSON, _ := json.Marshal(result)

	hasher.Write(inputJSON)
	hasher.Write(resultJSON)
	hasher.Write([]byte(policyID))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}
