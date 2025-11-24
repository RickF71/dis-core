package receipts

import (
	"context"

	"dis-core/internal/contextx"
)

// WithOriginDomain attaches the origin domain into the context so downstream
// emitters can resolve it. The value is stored as an untyped interface{} to
// avoid import cycles; callers may pass a *domain.Domain if they wish.
func WithOriginDomain(ctx context.Context, d any) context.Context {
	return contextx.WithOriginDomain(ctx, d)
}

// OriginDomainFromContext returns the raw origin value if present. Callers
// should type-assert to the concrete domain type as needed.
func OriginDomainFromContext(ctx context.Context) (any, bool) {
	return contextx.OriginDomainFromContext(ctx)
}
