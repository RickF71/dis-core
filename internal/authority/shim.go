package authority

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/authority/mutation"
	"dis-core/internal/identity"
)

// Compatibility shims for legacy internal/authority usage.
// These are minimal placeholders to keep the build green while logic
// migrates into internal/core/authority during MX-2.x. They intentionally
// provide only the surface needed by remaining callers and are safe to
// remove as migration progresses.

// SeatMutationEngine is a minimal placeholder for the real mutation engine.
type SeatMutationEngine struct{}

// SeatTransitionRequest represents a single seat transition request (placeholder).
type SeatTransitionRequest struct {
	// real fields exist in full implementation; kept empty for shim
}

// SeatTransitionBatchRequest represents a batch transition request (placeholder).
type SeatTransitionBatchRequest struct {
	Requests []SeatTransitionRequest
}

// SeatTransitionResult is a minimal result returned by mutation operations.
type SeatTransitionResult struct {
	OK bool `json:"ok"`
}

// MutateSeat performs a single seat mutation (shimbed, no-op).
func (e *SeatMutationEngine) MutateSeat(ctx context.Context, req SeatTransitionRequest) (*SeatTransitionResult, error) {
	return &SeatTransitionResult{OK: true}, nil
}

// MutateSeatBatch performs a batch of seat mutations (shimbed, no-op).
func (e *SeatMutationEngine) MutateSeatBatch(ctx context.Context, req SeatTransitionBatchRequest) (*SeatTransitionResult, error) {
	return &SeatTransitionResult{OK: true}, nil
}

// AuthorityDirection represents the direction of authority flow (placeholder).
type AuthorityDirection string

const (
	AuthorityDirectionUpward   AuthorityDirection = "upward"
	AuthorityDirectionDownward AuthorityDirection = "downward"
	AuthorityDirectionLateral  AuthorityDirection = "lateral"
)

// EvaluateResult is the result of an authority flow evaluation.
type EvaluateResult struct {
	Allow  bool
	Reason string
	Errors []string
}

// EvaluateAuthority is a shim for the real flow evaluation logic.
func EvaluateAuthority(action, seatState string, direction AuthorityDirection, actionDomain, seatDomain string, parentApproved, seatFrozen bool) *EvaluateResult {
	return &EvaluateResult{Allow: true, Reason: "shim-allow", Errors: nil}
}

// Minimal types referenced across authority API packages.
type AuthoritySchema struct {
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Schema      map[string]interface{} `json:"schema"`
}

// LoadSchema reads a JSON schema from disk into memory (compat shim).
func LoadSchema(path string) (*AuthoritySchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s AuthoritySchema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

type Console struct {
	LastVerification time.Time
	Domain           string
	CoreVersion      string
}

// NewConsole constructs a minimal Console instance (compat shim).
func NewConsole(db *pgxpool.Pool, led interface{}, triad *identity.TriadRepository, schema *AuthoritySchema) *Console {
	_ = db
	_ = led
	_ = triad
	_ = schema
	return &Console{LastVerification: time.Now(), Domain: "", CoreVersion: "shim"}
}

// EvaluateAuthority evaluates an input payload against the console schema.
// This is a shim that always allows; real logic should be migrated to core/authority.
func (c *Console) EvaluateAuthority(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	_ = ctx
	// Return a permissive evaluation result for compatibility.
	return map[string]interface{}{"allow": true, "reason": "shim-allow"}, nil
}

// DecisionFilter includes fields used by debug/list handlers.
type DecisionFilter struct {
	Domain   string
	PolicyID string
	Limit    int
	Offset   int
}

// Decision represents a policy decision record (minimal fields used by API handlers).
type Decision struct {
	ID         string                 `json:"id,omitempty"`
	Actor      string                 `json:"actor,omitempty"`
	Domain     string                 `json:"domain,omitempty"`
	PolicyID   string                 `json:"policy_id,omitempty"`
	Input      map[string]interface{} `json:"input,omitempty"`
	Result     map[string]interface{} `json:"result,omitempty"`
	Reason     string                 `json:"reason,omitempty"`
	PhaseTag   string                 `json:"phase_tag,omitempty"`
	ReplayHash string                 `json:"replay_hash,omitempty"`
}

// ListDecisions returns an empty list and no error in the shim.
func ListDecisions(ctx context.Context, db *pgxpool.Pool, filter DecisionFilter) ([]Decision, error) {
	_ = ctx
	_ = db
	_ = filter
	return []Decision{}, nil
}

// GetDecision returns a shimbed decision (not persisted).
func GetDecision(ctx context.Context, db *pgxpool.Pool, id string) (*Decision, error) {
	// Return a minimal Decision object to satisfy handlers during migration.
	return &Decision{ID: id, Domain: "unknown", Actor: "system", PolicyID: "", Result: map[string]interface{}{}}, nil
}

// StoreDecision persists a decision (shim: no-op).
func StoreDecision(ctx context.Context, db *pgxpool.Pool, d *Decision) error {
	_ = ctx
	_ = db
	_ = d
	return nil
}

// RecordAuthorityReceipt creates an immutable receipt for a governance action.
// Shimbed implementation: no-op
func RecordAuthorityReceipt(ctx context.Context, db *pgxpool.Pool, domainID interface{}, action string, payload map[string]interface{}, opts interface{}) error {
	_ = ctx
	_ = db
	_ = domainID
	_ = action
	_ = payload
	_ = opts
	return nil
}

// EventBus interface for SSE and event publishing (compat shim).
type EventBus interface {
	SSEHandler(w http.ResponseWriter, r *http.Request)
}

// in-memory event bus implementation (minimal)
type inMemoryEventBus struct{}

func (b *inMemoryEventBus) SSEHandler(w http.ResponseWriter, r *http.Request) {
	// Simple placeholder: return 204 No Content for event stream endpoints.
	w.WriteHeader(http.StatusNoContent)
}

// NewMemoryBus returns a shimbed in-memory event bus.
func NewMemoryBus() EventBus { return &inMemoryEventBus{} }

// NewSeatMutationEngine constructs a shimbed SeatMutationEngine for wiring.
func NewSeatMutationEngine(triadRepo *identity.TriadRepository, policy mutation.PolicyEvaluator, audit mutation.AuditLogger, eb EventBus) *SeatMutationEngine {
	_ = triadRepo
	_ = policy
	_ = audit
	_ = eb
	return &SeatMutationEngine{}
}
