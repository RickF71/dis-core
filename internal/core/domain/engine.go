package domain

import (
	"context"
	"fmt"

	"dis-core/internal/core/types"
	"dis-core/internal/receipts"
)

// Engine is the central orchestrator for all domain-scoped operations.
// Use generic fields (interface{}) to avoid import cycles during initial
// wiring; later phases can tighten these to concrete interfaces.
type Engine struct {
	DB       types.DBStore
	Policy   types.PolicyEngine
	Receipts types.ReceiptsStore
}

// NewEngine wires the dependencies.
func NewEngine(dbStore types.DBStore, p types.PolicyEngine, r types.ReceiptsStore) *Engine {
	return &Engine{
		DB:       dbStore,
		Policy:   p,
		Receipts: r,
	}
}

// ResolveDomain loads a domain by id.
func (e *Engine) ResolveDomain(ctx context.Context, id string) (*Domain, error) {
	dom, err := LoadDomainByID(ctx, e.DB, id)
	if err != nil {
		return nil, fmt.Errorf("resolve domain %s: %w", id, err)
	}
	return dom, nil
}

// HandleAction is the high-level entry point for all domain-scoped behavior.
// This is a skeleton; actual logic will be added in later phases.
func (e *Engine) HandleAction(
	ctx context.Context,
	dom *Domain,
	action string,
	payload map[string]any,
) error {
	// ----------------------------------------------------
	// A. Resolve actor and attach origin domain to context
	// ----------------------------------------------------
	actorID, _ := ctx.Value("actor_id").(string)
	if actorID == "" {
		return fmt.Errorf("missing actor_id in context")
	}

	// attach domain as origin (stored as untyped value)
	ctx = receipts.WithOriginDomain(ctx, dom)

	// ----------------------------------------------------
	// B. Evaluate domain-scoped policy
	// ----------------------------------------------------
	ns := fmt.Sprintf("dis.policy.domain.%s", dom.Name)
	decisionMap, err := e.Policy.EvalFn(ctx, ns, map[string]any{
		"action":  action,
		"payload": payload,
		"actor":   actorID,
		"domain":  dom.Name,
	})
	if err != nil {
		return fmt.Errorf("policy evaluation failed: %w", err)
	}

	// ----------------------------------------------------
	// C. Create the envelope
	// ----------------------------------------------------
	env := receipts.NewEnvelope(dom.ID, dom.Name, actorID)

	// ----------------------------------------------------
	// D. Populate panels
	// ----------------------------------------------------
	env.ActionPanel["verb"] = action
	env.ActionPanel["inputs"] = payload

	for k, v := range decisionMap {
		env.PolicyPanel[k] = v
	}

	env.DomainPanel["domain_id"] = dom.ID
	env.DomainPanel["domain_name"] = dom.Name

	// dimension string – use placeholder if not present
	if domDimension, ok := any("").(interface{}); ok { // placeholder
		_ = domDimension
	}
	env.DimensionPanel["dimension"] = "unknown"

	env.IdentityPanel["actor_id"] = actorID

	// ----------------------------------------------------
	// E. Persist envelope through ReceiptsStore
	// ----------------------------------------------------
	if err := e.Receipts.SaveEnvelope(ctx, env); err != nil {
		return fmt.Errorf("save envelope: %w", err)
	}

	// ----------------------------------------------------
	// F. Dispatch to future per-action handlers (stub)
	// ----------------------------------------------------
	return nil
}

// Placeholder / thin façade; not yet used.
func LoadDomainByID(ctx context.Context, store types.DBStore, id string) (*Domain, error) {
	domRaw, err := store.LoadDomainByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dom, ok := domRaw.(*Domain)
	if !ok {
		return nil, fmt.Errorf("invalid domain type returned by DBStore")
	}
	return dom, nil
}
