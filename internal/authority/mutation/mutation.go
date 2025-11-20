package mutation

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// PolicyEvaluator is an interface for evaluating policies (shim).
type PolicyEvaluator interface{}

// AuditLogger is an interface for audit logging (shim).
type AuditLogger interface{}

// NewOPAPolicyEvaluator returns a shimbed policy evaluator.
func NewOPAPolicyEvaluator(bundleDir string, verbose bool) (PolicyEvaluator, error) {
	// Real implementation will load OPA bundles; shim returns nil evaluator.
	return nil, nil
}

// NullPolicyEvaluator is a no-op evaluator used when policy backend is disabled.
type NullPolicyEvaluator struct{}

// NewDBAuditLogger returns a shimbed DB audit logger.
// Signature matches the expectations in wiring_gov4.go: (db *pgxpool.Pool, debug bool, receiptKind, table string) AuditLogger
func NewDBAuditLogger(db *pgxpool.Pool, debug bool, receiptKind, table string) AuditLogger {
	// Shim: return nil implementation (caller treats as interface)
	_ = db
	_ = debug
	_ = receiptKind
	_ = table
	return nil
}

// NullAuditLogger placeholder
type NullAuditLogger struct{}
