package domain

import "context"

// ActionContext carries actor and inputs for Engine.Do convenience facade.
type ActionContext struct {
	ActorID string
	Inputs  map[string]any
}

// Do attaches actor information to the context and delegates to HandleAction.
func (e *Engine) Do(
	ctx context.Context,
	dom *Domain,
	verb string,
	ac ActionContext,
) error {
	ctx = context.WithValue(ctx, "actor_id", ac.ActorID)
	return e.HandleAction(ctx, dom, verb, ac.Inputs)
}
