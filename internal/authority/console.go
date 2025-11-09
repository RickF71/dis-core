package authority

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dis-core/internal/ledger"
	"dis-core/internal/policy"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Console is the runtime authority subsystem and carries descriptive metadata.
type Console struct {
	DB     *pgxpool.Pool
	Ledger *ledger.Ledger
	Policy policy.PolicyEngine
	Schema *AuthoritySchema

	// --- added metadata fields ---
	Domain           string    `json:"domain"`
	CoreVersion      string    `json:"core_version"`
	LastVerification time.Time `json:"last_verification"`
}

// NewConsole creates a new Authority Console.
func NewConsole(db *pgxpool.Pool, led *ledger.Ledger, pe policy.PolicyEngine, schema *AuthoritySchema) *Console {
	return &Console{
		DB:               db,
		Ledger:           led,
		Policy:           pe,
		Schema:           schema,
		Domain:           "domain.terra",  // default placeholder
		CoreVersion:      "DIS-CORE v1.0", // version tag for reporting
		LastVerification: time.Time{},     // will update dynamically
	}
}

// EvaluateAuthority executes an action through the policy engine
// and records a receipt in the ledger if authorized or denied.
func (c *Console) EvaluateAuthority(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if c.Policy == nil {
		return nil, fmt.Errorf("policy engine not initialized")
	}

	decision, err := c.Policy.EvaluateAction(ctx, input)
	if err != nil {
		return nil, err
	}

	// --------------------------------------
	// 📜 Record full decision as JSON receipt
	// --------------------------------------
	if c.Ledger != nil {
		payload, err := json.Marshal(map[string]interface{}{
			"decision": decision,
			"input":    input,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal receipt payload: %w", err)
		}

		rcpt := ledger.Receipt{
			Type:      "ci.call.v1",
			Actor:     fmt.Sprint(input["actor"]),
			Target:    fmt.Sprint(input["target"]),
			Domain:    fmt.Sprint(input["domain"]),
			Payload:   payload,
			CreatedAt: time.Now().UTC(),
		}

		if err := c.Ledger.InsertReceipt(ctx, rcpt); err != nil {
			return nil, fmt.Errorf("failed to log receipt: %v", err)
		}
	}

	// Update last verification timestamp (if applicable)
	c.LastVerification = time.Now().UTC()

	return map[string]interface{}{
		"decision": decision,
	}, nil
}
