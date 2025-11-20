package auth

import (
	"context"
	"net/http"
)

// ActiveUser represents the current user's authentication state.
// External UID is never persisted or logged - it only exists in request context.
type ActiveUser struct {
	// ExternalUID is the raw external identifier (e.g., from OAuth, OIDC)
	// NEVER stored in database, NEVER appears in receipts or logs
	ExternalUID string

	// HasExternalUID indicates if an external UID was provided (even if not bound)
	HasExternalUID bool

	// Bound indicates if the external UID has been mapped to a corporeal domain
	Bound bool

	// CorporealDomainID is the internal ID of the corporeal domain (if bound)
	CorporealDomainID int64

	// CorporealDomainUID is the sovereign DIS identity (if bound)
	CorporealDomainUID string
}

// contextKey is a private type for context keys to avoid collisions
type contextKey string

const activeUserKey contextKey = "activeUser"

// WithActiveUser attaches an ActiveUser to a request context
func WithActiveUser(ctx context.Context, user *ActiveUser) context.Context {
	return context.WithValue(ctx, activeUserKey, user)
}

// GetActiveUser retrieves the ActiveUser from a request context
// Returns nil if no user is attached
func GetActiveUser(r *http.Request) *ActiveUser {
	user, ok := r.Context().Value(activeUserKey).(*ActiveUser)
	if !ok {
		return nil
	}
	return user
}

// GetActiveUserFromCtx retrieves the ActiveUser from a context.Context directly.
// This is useful for non-HTTP layers (engines) that only have a context.
func GetActiveUserFromCtx(ctx context.Context) *ActiveUser {
	user, ok := ctx.Value(activeUserKey).(*ActiveUser)
	if !ok {
		return nil
	}
	return user
}

// IsAuthenticated returns true if an external UID is present
func (u *ActiveUser) IsAuthenticated() bool {
	return u != nil && u.HasExternalUID
}

// IsBound returns true if the user is authenticated AND bound to a corporeal domain
func (u *ActiveUser) IsBound() bool {
	return u != nil && u.Bound && u.CorporealDomainID > 0
}

// Active Actor context management
// contextKeyActiveActor is the key for storing the active actor seat ID
const activeActorKey contextKey = "activeActor"

// SetActiveActor stores the active actor seat ID in the request context
// Returns a new request with the updated context
func SetActiveActor(r *http.Request, seatID string) *http.Request {
	ctx := context.WithValue(r.Context(), activeActorKey, seatID)
	return r.WithContext(ctx)
}

// GetActiveActor retrieves the active actor seat ID from the request context
// Returns the seat ID and a boolean indicating if an active actor is set
func GetActiveActor(r *http.Request) (string, bool) {
	seatID, ok := r.Context().Value(activeActorKey).(string)
	return seatID, ok
}
