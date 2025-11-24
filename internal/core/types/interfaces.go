package types

import (
	"context"
)

// DBStore defines minimal DB operations used by the Domain Engine.
type DBStore interface {
	LoadDomainByID(ctx context.Context, id string) (any, error)
}

// PolicyEngine defines evaluation and decision contract.
type PolicyEngine interface {
	// EvalFn evaluates a policy namespace with input and returns a generic
	// decision map (implementation-defined) and an error.
	EvalFn(ctx context.Context, namespace string, input map[string]any) (map[string]any, error)
}

// ReceiptsStore defines persistence for envelopes.
type ReceiptsStore interface {
	// SaveEnvelope persists an envelope; use a generic env (implementation
	// decides concrete type) to avoid import cycles.
	SaveEnvelope(ctx context.Context, env any) error
}
