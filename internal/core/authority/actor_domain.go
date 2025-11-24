package authority

import (
	"context"

	"dis-core/internal/contextx"
	cdomain "dis-core/internal/core/domain"
)

// ResolveActorDomain returns the domain in which the current actor is
// performing the action. It first prefers a domain attached to the context
// (set by API middleware). If none is present it falls back to an empty
// domain object; callers should handle empty IDs appropriately.
func ResolveActorDomain(ctx context.Context) *cdomain.Domain {
	if v, ok := contextx.OriginDomainFromContext(ctx); ok {
		if d, ok2 := v.(*cdomain.Domain); ok2 {
			return d
		}
	}
	return &cdomain.Domain{}
}
