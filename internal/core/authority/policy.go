package authority

import (
	"context"
	"fmt"

	authpkg "dis-core/internal/auth"
	contextx "dis-core/internal/contextx"
	dbstore "dis-core/internal/db"

	"github.com/google/uuid"
)

// activeUser helper returns the ActiveUser from the context (or nil).
func (e *Engine) activeUser(ctx context.Context) *authpkg.ActiveUser {
	return authpkg.GetActiveUserFromCtx(ctx)
}

// CanInstantiateSeat checks whether the caller (via ctx) and contract allow
// instantiation of a seat in domainID for identityID. contractID may be empty.
// Returns (allowed, nil) for allowed, (false, nil) for a policy denial, and
// (false, err) for system/policy errors.
func (e *Engine) CanInstantiateSeat(ctx context.Context, domainID, identityID, contractID string) (bool, error) {
	usr := e.activeUser(ctx)
	if usr == nil {
		// unauthenticated callers cannot instantiate
		return false, nil
	}

	// If DB not configured, fallback to deny (but signal no error)
	if e.DB == nil {
		return false, nil
	}

	// Check domain freeze state: deny write actions if a write or all freeze is active.
	if denied, _, _ := e.checkDomainFreeze(ctx, domainID, "write"); denied {
		// soft-deny: do not surface a system error
		return false, nil
	}

	// Rule 1: If the caller is bound to the same corporeal identity (self-assert), allow
	if usr.CorporealDomainUID != "" && usr.CorporealDomainUID == identityID {
		return true, nil
	}

	// Rule 2: If the caller has a root/admin seat in the target domain, allow
	seats, err := dbstore.ListSeatsByDomain(ctx, e.DB, domainID)
	if err != nil {
		return false, fmt.Errorf("policy: failed to list seats for domain: %w", err)
	}
	for _, s := range seats {
		if mid, ok := s["identity_id"].(string); ok && mid != "" {
			if mid == usr.CorporealDomainUID {
				if st, ok := s["kind"].(string); ok && (st == "root" || st == "member") {
					// Evaluate AT-1 example policy via engine EvalFn callback before allowing
					if e.EvalFn != nil {
						attrs := map[string]interface{}{"context": map[string]interface{}{"attrs": map[string]interface{}{"corrupt": false}}}
						allowed, _, details, err := e.EvalFn(ctx, attrs)
						if err != nil {
							return false, fmt.Errorf("policy eval error: %w", err)
						}
						// Propagate policy decision details into the context so
						// downstream receipt emitters can pick them up for provenance.
						if details != nil {
							ctx = contextx.WithPolicyDecisionMap(ctx, details)
						}
						if !allowed {
							return false, nil
						}
					}
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// CanViewSeat determines if the caller in ctx may view the given seat.
// Returns (false, nil) for soft-deny visibility (caller should see nothing).
func (e *Engine) CanViewSeat(ctx context.Context, seatID string) (bool, error) {
	usr := e.activeUser(ctx)
	if usr == nil {
		return false, nil
	}
	if e.DB == nil {
		return false, nil
	}

	// Fetch seat meta
	seat, err := dbstore.GetSeat(ctx, e.DB, seatID)
	if err != nil {
		return false, fmt.Errorf("policy: failed to fetch seat meta: %w", err)
	}
	if seat == nil {
		return false, nil
	}

	// Check domain-level freeze: deny if 'all' freeze present for the seat's domain.
	if did, ok := seat["domain_id"].(string); ok && did != "" {
		if denied, _, _ := e.checkDomainFreeze(ctx, did, ""); denied {
			return false, nil
		}
	}

	// Self-ownership: ActiveUser bound identity matches seat identity
	if sid, ok := seat["identity_id"].(string); ok && sid != "" {
		if usr.CorporealDomainUID != "" && usr.CorporealDomainUID == sid {
			return true, nil
		}
	}

	// Domain admin membership
	if did, ok := seat["domain_id"].(string); ok && did != "" {
		seats, err := dbstore.ListSeatsByDomain(ctx, e.DB, did)
		if err != nil {
			return false, fmt.Errorf("policy: failed to list seats for domain: %w", err)
		}
		for _, s := range seats {
			if mid, ok := s["identity_id"].(string); ok && mid != "" {
				if mid == usr.CorporealDomainUID {
					if st, ok := s["kind"].(string); ok && (st == "root") {
						return true, nil
					}
				}
			}
		}
	}

	// Default deny (soft)
	return false, nil
}

// CanListDomainSeats checks whether the caller can list seats in domainID.
func (e *Engine) CanListDomainSeats(ctx context.Context, domainID string) (bool, error) {
	usr := e.activeUser(ctx)
	if usr == nil {
		return false, nil
	}
	if e.DB == nil {
		return false, nil
	}

	// Check domain freeze: listing is allowed unless an 'all' freeze exists.
	if denied, _, _ := e.checkDomainFreeze(ctx, domainID, ""); denied {
		return false, nil
	}

	seats, err := dbstore.ListSeatsByDomain(ctx, e.DB, domainID)
	if err != nil {
		return false, fmt.Errorf("policy: failed to list seats for domain: %w", err)
	}
	for _, s := range seats {
		if mid, ok := s["identity_id"].(string); ok && mid != "" {
			if mid == usr.CorporealDomainUID {
				if st, ok := s["kind"].(string); ok && (st == "root" || st == "member") {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// CanViewSeatLineage checks whether the caller may view the lineage for a seat
// associated with domainID and identityID. Returns (false,nil) for soft-deny.
func (e *Engine) CanViewSeatLineage(ctx context.Context, domainID, identityID string) (bool, error) {
	usr := e.activeUser(ctx)
	if usr == nil {
		return false, nil
	}
	if e.DB == nil {
		return false, nil
	}

	// Check for domain-level all-scope freeze before lineage checks
	if denied, _, _ := e.checkDomainFreeze(ctx, domainID, ""); denied {
		return false, nil
	}

	// Self-ownership
	if usr.CorporealDomainUID != "" && usr.CorporealDomainUID == identityID {
		return true, nil
	}

	// Domain admin check
	seats, err := dbstore.ListSeatsByDomain(ctx, e.DB, domainID)
	if err != nil {
		return false, fmt.Errorf("policy: failed to list seats for domain: %w", err)
	}
	for _, s := range seats {
		if mid, ok := s["identity_id"].(string); ok && mid != "" {
			if mid == usr.CorporealDomainUID {
				if st, ok := s["kind"].(string); ok && (st == "root") {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// checkDomainFreeze inspects active freezes for domainID and returns (denied, reason, err).
// It returns a soft-deny (denied=true, err=nil) on transient DB errors to avoid panics.
func (e *Engine) checkDomainFreeze(ctx context.Context, domainID, scope string) (bool, string, error) {
	if e.DB == nil {
		return false, "", nil
	}
	uid, err := uuid.Parse(domainID)
	if err != nil {
		return false, "", nil
	}
	frs, err := e.GetActiveFreezes(ctx, uid)
	if err != nil {
		// soft-deny on DB failures
		return true, "deny:freeze:sys", nil
	}
	for _, fr := range frs {
		if fr.Scope == "all" {
			return true, "deny:freeze:all", nil
		}
		if scope != "" && fr.Scope == scope {
			return true, fmt.Sprintf("deny:freeze:%s", scope), nil
		}
	}
	return false, "", nil
}
