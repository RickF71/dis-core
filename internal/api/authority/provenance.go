package authorityapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/authority"
)

// ProvenanceInfo captures the audit trail for authority decisions
type ProvenanceInfo struct {
	ReceiptID    string    `json:"receipt_id"`
	SourceChain  []string  `json:"source_chain"`
	RequestID    string    `json:"request_id"`
	UserAgent    string    `json:"user_agent,omitempty"`
	RemoteAddr   string    `json:"remote_addr,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	PolicyEngine string    `json:"policy_engine"`
}

// CreateProvenanceReceipt creates a receipt for authority decisions
func CreateProvenanceReceipt(ctx context.Context, db *pgxpool.Pool, decision *authority.Decision, provenance ProvenanceInfo) error {
	// Create receipt entry
	receiptID := uuid.New().String()

	receipt := map[string]interface{}{
		"id":         receiptID,
		"type":       "authority.decision.v1",
		"actor":      decision.Actor,
		"target":     decision.Domain,
		"domain":     decision.Domain,
		"created_at": time.Now(),
		"payload": map[string]interface{}{
			"decision_id": decision.ID,
			"policy_id":   decision.PolicyID,
			"input":       decision.Input,
			"result":      decision.Result,
			"reason":      decision.Reason,
			"replay_hash": decision.ReplayHash,
			"provenance":  provenance,
			"phase_tag":   decision.PhaseTag,
		},
	}

	// Store receipt in database
	query := `
		INSERT INTO receipts (id, type, actor, target, domain, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	payloadJSON, err := json.Marshal(receipt["payload"])
	if err != nil {
		return fmt.Errorf("failed to marshal receipt payload: %w", err)
	}

	_, err = db.Exec(ctx, query,
		receiptID,
		receipt["type"],
		receipt["actor"],
		receipt["target"],
		receipt["domain"],
		payloadJSON,
		receipt["created_at"],
	)

	if err != nil {
		return fmt.Errorf("failed to store provenance receipt: %w", err)
	}

	// Update decision with receipt link
	updateQuery := `
		UPDATE authority_decisions
		SET input = input || jsonb_build_object('receipt_id', $2)
		WHERE id = $1
	`

	_, err = db.Exec(ctx, updateQuery, decision.ID, receiptID)
	if err != nil {
		return fmt.Errorf("failed to link receipt to decision: %w", err)
	}

	return nil
}

// GetDecisionProvenance retrieves the full provenance trail for a decision
func GetDecisionProvenance(ctx context.Context, db *pgxpool.Pool, decisionID string) (map[string]interface{}, error) {
	// Get decision
	decision, err := authority.GetDecision(ctx, db, decisionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	// Get linked receipt
	var receiptID string
	if receipt, ok := decision.Input["receipt_id"]; ok {
		receiptID = receipt.(string)
	} else {
		return map[string]interface{}{
			"decision": decision,
			"receipt":  nil,
			"status":   "no_receipt",
		}, nil
	}

	// Query receipt
	receiptQuery := `
		SELECT id, type, actor, target, domain, payload, created_at
		FROM receipts
		WHERE id = $1
	`

	var receipt struct {
		ID        string                 `json:"id"`
		Type      string                 `json:"type"`
		Actor     string                 `json:"actor"`
		Target    string                 `json:"target"`
		Domain    string                 `json:"domain"`
		Payload   map[string]interface{} `json:"payload"`
		CreatedAt time.Time              `json:"created_at"`
	}

	var payloadJSON []byte
	err = db.QueryRow(ctx, receiptQuery, receiptID).Scan(
		&receipt.ID, &receipt.Type, &receipt.Actor,
		&receipt.Target, &receipt.Domain, &payloadJSON, &receipt.CreatedAt,
	)

	if err != nil {
		return map[string]interface{}{
			"decision":   decision,
			"receipt":    nil,
			"status":     "receipt_missing",
			"receipt_id": receiptID,
		}, nil
	}

	// Unmarshal payload
	err = json.Unmarshal(payloadJSON, &receipt.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal receipt payload: %w", err)
	}

	return map[string]interface{}{
		"decision": decision,
		"receipt":  receipt,
		"status":   "complete",
	}, nil
}

// LinkDecisionToChain creates a source chain link between decisions
func LinkDecisionToChain(ctx context.Context, db *pgxpool.Pool, decisionID, parentDecisionID string) error {
	// Update decision with parent link
	updateQuery := `
		UPDATE authority_decisions
		SET input = input || jsonb_build_object('parent_decision_id', $2)
		WHERE id = $1
	`

	_, err := db.Exec(ctx, updateQuery, decisionID, parentDecisionID)
	return err
}

// GetDecisionChain retrieves the full decision chain for audit
func GetDecisionChain(ctx context.Context, db *pgxpool.Pool, decisionID string) ([]*authority.Decision, error) {
	var chain []*authority.Decision
	currentID := decisionID

	for currentID != "" {
		decision, err := authority.GetDecision(ctx, db, currentID)
		if err != nil {
			break
		}

		chain = append(chain, decision)

		// Look for parent decision
		if parent, ok := decision.Input["parent_decision_id"]; ok {
			currentID = parent.(string)
		} else {
			break
		}

		// Prevent infinite loops
		if len(chain) > 100 {
			break
		}
	}

	return chain, nil
}
