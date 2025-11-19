package api

import (
	"context"
	"encoding/json"
	"net/http"

	"dis-core/internal/identity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GOV-11C: Identity Projections API
// Provides read-only access to domain-scoped identity projections and foreign acceptances

// IdentityProjection represents a domain's view of an actor's identity
type IdentityProjection struct {
	DomainID           uuid.UUID                   `json:"domain_id"`
	DomainName         string                      `json:"domain_name"`
	LocalIdentity      *string                     `json:"local_identity,omitempty"`      // domain.id value
	AcceptedIdentities []ForeignIdentityAcceptance `json:"accepted_identities,omitempty"` // foreign IDs
	ReceiptCount       int                         `json:"receipt_count"`
	IntegrityValid     bool                        `json:"integrity_valid"`
	LastActivity       string                      `json:"last_activity"`
}

// ForeignIdentityAcceptance represents an accepted external identity
type ForeignIdentityAcceptance struct {
	SourceDomainID   uuid.UUID `json:"source_domain_id"`
	SourceDomainName string    `json:"source_domain_name,omitempty"`
	ExternalSubject  string    `json:"external_subject"`
	Scope            *string   `json:"scope,omitempty"`
	ReceiptID        uuid.UUID `json:"receipt_id"`
	AcceptedAt       string    `json:"accepted_at"`
	RevokedAt        *string   `json:"revoked_at,omitempty"`
	Active           bool      `json:"active"`
}

// IdentityProjectionsSummary aggregates all projections for an actor
type IdentityProjectionsSummary struct {
	ActorID         uuid.UUID            `json:"actor_id"`
	Projections     []IdentityProjection `json:"projections"`
	TotalDomains    int                  `json:"total_domains"`
	TotalReceipts   int                  `json:"total_receipts"`
	IntegrityStatus string               `json:"integrity_status"` // "all-valid", "some-broken", "no-data"
}

// handleGetIdentityProjections returns all domain.id projections and foreign acceptances for an actor
func (s *Server) handleGetIdentityProjections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	actorIDStr := chi.URLParam(r, "actor_id")
	actorID, err := uuid.Parse(actorIDStr)
	if err != nil {
		http.Error(w, "invalid actor_id format", http.StatusBadRequest)
		return
	}

	summary, err := s.buildIdentityProjectionsSummary(ctx, actorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// buildIdentityProjectionsSummary constructs the projections summary from identity receipts
func (s *Server) buildIdentityProjectionsSummary(ctx context.Context, actorID uuid.UUID) (*IdentityProjectionsSummary, error) {
	// Get full lineage
	lineage, err := identity.GetIdentityLineage(ctx, s.db, actorID)
	if err != nil {
		return nil, err
	}

	// Build domain-keyed map of projections
	projectionMap := make(map[uuid.UUID]*IdentityProjection)
	var totalReceipts int

	for _, entry := range lineage.Entries {
		totalReceipts++

		// Initialize projection for this domain if not exists
		if _, exists := projectionMap[entry.DomainID]; !exists {
			projectionMap[entry.DomainID] = &IdentityProjection{
				DomainID:       entry.DomainID,
				ReceiptCount:   0,
				IntegrityValid: true,
			}
		}

		proj := projectionMap[entry.DomainID]
		proj.ReceiptCount++
		proj.LastActivity = entry.CreatedAt

		if !entry.Valid {
			proj.IntegrityValid = false
		}

		// Handle action-specific logic
		switch entry.Action {
		case string(identity.IdentityDomainIDCreateV1):
			// Extract domain.id value from payload
			if domainIDVal, ok := entry.Payload["domain_identity_value"].(string); ok {
				proj.LocalIdentity = &domainIDVal
			}

		case string(identity.IdentityAcceptV1):
			// Add foreign acceptance
			if entry.SourceDomainID != nil && entry.ExternalSubject != nil {
				acceptance := ForeignIdentityAcceptance{
					SourceDomainID:  *entry.SourceDomainID,
					ExternalSubject: *entry.ExternalSubject,
					Scope:           entry.Scope,
					ReceiptID:       entry.ID,
					AcceptedAt:      entry.CreatedAt,
					Active:          true,
				}
				proj.AcceptedIdentities = append(proj.AcceptedIdentities, acceptance)
			}

		case string(identity.IdentityAcceptRevokeV1):
			// Mark foreign acceptance as revoked
			if entry.ExternalSubject != nil {
				for i := range proj.AcceptedIdentities {
					if proj.AcceptedIdentities[i].ExternalSubject == *entry.ExternalSubject {
						proj.AcceptedIdentities[i].Active = false
						proj.AcceptedIdentities[i].RevokedAt = &entry.CreatedAt
					}
				}
			}
		}
	}

	// Convert map to slice
	var projections []IdentityProjection
	for _, proj := range projectionMap {
		projections = append(projections, *proj)
	}

	// Determine overall integrity status
	integrityStatus := "all-valid"
	if totalReceipts == 0 {
		integrityStatus = "no-data"
	} else {
		for _, proj := range projections {
			if !proj.IntegrityValid {
				integrityStatus = "some-broken"
				break
			}
		}
	}

	return &IdentityProjectionsSummary{
		ActorID:         actorID,
		Projections:     projections,
		TotalDomains:    len(projections),
		TotalReceipts:   totalReceipts,
		IntegrityStatus: integrityStatus,
	}, nil
}

// handleGetDomainMemberIdentity returns domain-specific identity view for a member
func (s *Server) handleGetDomainMemberIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	domainIDStr := chi.URLParam(r, "domain_id")
	actorIDStr := chi.URLParam(r, "actor_id")

	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		http.Error(w, "invalid domain_id format", http.StatusBadRequest)
		return
	}

	actorID, err := uuid.Parse(actorIDStr)
	if err != nil {
		http.Error(w, "invalid actor_id format", http.StatusBadRequest)
		return
	}

	// Get full projections summary
	summary, err := s.buildIdentityProjectionsSummary(ctx, actorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter to just this domain
	var domainProjection *IdentityProjection
	for _, proj := range summary.Projections {
		if proj.DomainID == domainID {
			domainProjection = &proj
			break
		}
	}

	if domainProjection == nil {
		http.Error(w, "no identity projection found for this domain and actor", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(domainProjection); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
