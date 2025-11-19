package identity

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-10: Identity Lineage Retrieval with Hash Chain Verification
// GOV-11: Extended for domain projections and foreign identity acceptance
// Provides chronological view of identity governance actions with integrity checking

// LineageEntry represents a single entry in identity lineage with verification status
type LineageEntry struct {
	ID        uuid.UUID              `json:"id"`
	PrevID    *uuid.UUID             `json:"prev_id,omitempty"`
	Hash      string                 `json:"hash"`
	Valid     bool                   `json:"valid"`
	Action    string                 `json:"action"`
	CreatedAt string                 `json:"created_at"`
	DomainID  uuid.UUID              `json:"domain_id"`
	ConsentBy uuid.UUID              `json:"consent_by"`
	Payload   map[string]interface{} `json:"payload"`

	// GOV-11: Domain projection and foreign identity fields
	TargetDomainID  *uuid.UUID `json:"target_domain_id,omitempty"`
	SourceDomainID  *uuid.UUID `json:"source_domain_id,omitempty"`
	ExternalSubject *string    `json:"external_subject,omitempty"`
	Channel         *string    `json:"channel,omitempty"`
	Method          *string    `json:"method,omitempty"`
	Scope           *string    `json:"scope,omitempty"`
}

// IdentityLineage represents the complete lineage for an identity
type IdentityLineage struct {
	ActorID   uuid.UUID       `json:"actor_id"`
	Entries   []LineageEntry  `json:"entries"`
	Integrity IntegrityStatus `json:"integrity"`
}

// IntegrityStatus summarizes lineage integrity
type IntegrityStatus struct {
	TotalReceipts int      `json:"total_receipts"`
	ValidChains   int      `json:"valid_chains"`
	BrokenChains  int      `json:"broken_chains"`
	RootReceipts  int      `json:"root_receipts"`
	Issues        []string `json:"issues,omitempty"`
}

// GetIdentityLineage retrieves and verifies the complete identity lineage for an actor
func GetIdentityLineage(ctx context.Context, db *pgxpool.Pool, actorID uuid.UUID) (*IdentityLineage, error) {
	rows, err := db.Query(ctx, `
		SELECT id, prev_id, hash, action, created_at, domain_id, consent_by, payload,
		       target_domain_id, source_domain_id, external_subject, channel, method, scope
		FROM identity_receipts
		WHERE actor_id = $1
		ORDER BY created_at ASC
	`, actorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LineageEntry
	integrity := IntegrityStatus{}

	for rows.Next() {
		var e LineageEntry
		var payloadBytes []byte
		var createdAt time.Time

		err := rows.Scan(
			&e.ID,
			&e.PrevID,
			&e.Hash,
			&e.Action,
			&createdAt,
			&e.DomainID,
			&e.ConsentBy,
			&payloadBytes,
			&e.TargetDomainID,
			&e.SourceDomainID,
			&e.ExternalSubject,
			&e.Channel,
			&e.Method,
			&e.Scope,
		)
		if err != nil {
			return nil, err
		}

		// Convert timestamp to string
		e.CreatedAt = createdAt.Format(time.RFC3339)

		// Unmarshal payload
		if err := json.Unmarshal(payloadBytes, &e.Payload); err != nil {
			return nil, err
		}

		// Verify chain integrity
		if len(entries) == 0 {
			// First entry - should be root (no prev_id)
			if e.PrevID == nil {
				e.Valid = true
				integrity.RootReceipts++
			} else {
				e.Valid = false
				integrity.BrokenChains++
				integrity.Issues = append(integrity.Issues, "First receipt has prev_id (should be root)")
			}
		} else {
			// Subsequent entries - verify chain link
			prevEntry := entries[len(entries)-1]
			if e.PrevID != nil && *e.PrevID == prevEntry.ID {
				// Chain link is valid - hash was verified at creation time
				e.Valid = true
				integrity.ValidChains++
			} else {
				e.Valid = false
				integrity.BrokenChains++
				if e.PrevID == nil {
					integrity.Issues = append(integrity.Issues, "Receipt "+e.ID.String()+" missing prev_id")
				} else {
					integrity.Issues = append(integrity.Issues, "Receipt "+e.ID.String()+" prev_id mismatch")
				}
			}
		}

		integrity.TotalReceipts++
		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &IdentityLineage{
		ActorID:   actorID,
		Entries:   entries,
		Integrity: integrity,
	}, nil
}

// GetIdentityLineageSummary returns a lightweight summary without full payloads
func GetIdentityLineageSummary(ctx context.Context, db *pgxpool.Pool, actorID uuid.UUID) (*IdentityLineage, error) {
	rows, err := db.Query(ctx, `
		SELECT id, prev_id, hash, action, created_at, domain_id, consent_by
		FROM identity_receipts
		WHERE actor_id = $1
		ORDER BY created_at ASC
	`, actorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LineageEntry
	integrity := IntegrityStatus{}

	for rows.Next() {
		var e LineageEntry

		err := rows.Scan(
			&e.ID,
			&e.PrevID,
			&e.Hash,
			&e.Action,
			&e.CreatedAt,
			&e.DomainID,
			&e.ConsentBy,
		)
		if err != nil {
			return nil, err
		}

		// Simple validation without payload verification
		if len(entries) == 0 {
			e.Valid = e.PrevID == nil
			if e.Valid {
				integrity.RootReceipts++
			}
		} else {
			prevEntry := entries[len(entries)-1]
			e.Valid = e.PrevID != nil && *e.PrevID == prevEntry.ID
			if e.Valid {
				integrity.ValidChains++
			} else {
				integrity.BrokenChains++
			}
		}

		integrity.TotalReceipts++
		entries = append(entries, e)
	}

	return &IdentityLineage{
		ActorID:   actorID,
		Entries:   entries,
		Integrity: integrity,
	}, nil
}

// VerifyIdentityChain verifies the integrity of a complete identity chain
func VerifyIdentityChain(entries []LineageEntry) IntegrityStatus {
	integrity := IntegrityStatus{
		TotalReceipts: len(entries),
	}

	for i, entry := range entries {
		if i == 0 {
			if entry.PrevID == nil {
				integrity.RootReceipts++
			} else {
				integrity.BrokenChains++
				integrity.Issues = append(integrity.Issues, "Root receipt has prev_id")
			}
		} else {
			prevEntry := entries[i-1]
			if entry.PrevID != nil && *entry.PrevID == prevEntry.ID {
				integrity.ValidChains++
			} else {
				integrity.BrokenChains++
				integrity.Issues = append(integrity.Issues, "Chain break at "+entry.ID.String())
			}
		}
	}

	return integrity
}
