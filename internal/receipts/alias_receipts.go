package receipts

import (
	"context"
	"time"

	"dis-core/internal/domain"
	"dis-core/internal/identity"

	"github.com/google/uuid"
)

// GOV-12: Alias Canon & DSCI Integration - Receipt Helpers
// These functions provide convenience wrappers for recording alias-related receipts
// using the existing GOV-10 identity receipt system.

// RecordAliasAddReceipt records an identity.alias.add.v1 receipt
// for canonical alias creation (AUTO/RELATIONSHIP/MASK).
//
// Parameters:
// - ctx: request context
// - store: identity receipt store (GOV-10)
// - alias: the newly created alias
// - actorID: the seat/identity performing the action
// - consentBy: the seat providing consent (usually same as actor for self-service)
func RecordAliasAddReceipt(
	ctx context.Context,
	store *identity.IdentityReceiptStore,
	alias *domain.Alias,
	actorID uuid.UUID,
	consentBy uuid.UUID,
) (*identity.IdentityReceipt, error) {
	if store == nil {
		return nil, nil // Graceful degradation if receipt system not configured
	}

	// Build payload with canonical alias metadata
	payload := map[string]interface{}{
		"alias_id":         alias.ID.String(),
		"owner_domain_id":  alias.OwnerDomainID.String(),
		"target_domain_id": alias.TargetDomainID.String(),
		"alias_type":       string(alias.AliasType),
		"alias_name":       alias.AliasName,
		"created_at":       alias.CreatedAt.Format(time.RFC3339),
	}

	// Add optional fields
	if alias.IsCorporealAuto {
		payload["is_corporeal_auto"] = true
	}
	if alias.DSCIContractID != nil {
		payload["dsci_contract_id"] = alias.DSCIContractID.String()
	}
	if len(alias.Metadata) > 0 {
		payload["metadata"] = alias.Metadata
	}

	// Create receipt
	receipt := &identity.IdentityReceipt{
		DomainID:  alias.OwnerDomainID,
		ActorID:   actorID,
		Action:    identity.IdentityAliasAddV1,
		ConsentBy: consentBy,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}

	// Record in identity receipt ledger
	if err := store.RecordIdentityReceipt(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

// RecordAliasRetireReceipt records an identity.alias.remove.v1 receipt
// for alias retirement (setting retired_at).
//
// Parameters:
// - ctx: request context
// - store: identity receipt store (GOV-10)
// - aliasID: UUID of the alias being retired
// - ownerDomainID: domain that owns the alias
// - reason: reason for retirement (e.g., "user request", "TTL expired", "contract terminated")
// - actorID: the seat/identity performing the action
// - consentBy: the seat providing consent
func RecordAliasRetireReceipt(
	ctx context.Context,
	store *identity.IdentityReceiptStore,
	aliasID uuid.UUID,
	ownerDomainID uuid.UUID,
	reason string,
	actorID uuid.UUID,
	consentBy uuid.UUID,
) (*identity.IdentityReceipt, error) {
	if store == nil {
		return nil, nil // Graceful degradation if receipt system not configured
	}

	// Build payload
	payload := map[string]interface{}{
		"alias_id":   aliasID.String(),
		"retired_at": time.Now().UTC().Format(time.RFC3339),
		"reason":     reason,
	}

	// Create receipt
	receipt := &identity.IdentityReceipt{
		DomainID:  ownerDomainID,
		ActorID:   actorID,
		Action:    identity.IdentityAliasRemoveV1,
		ConsentBy: consentBy,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}

	// Record in identity receipt ledger
	if err := store.RecordIdentityReceipt(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

// AliasReceiptPayload represents the structured payload for alias receipts
// (for documentation and validation purposes)
type AliasReceiptPayload struct {
	// Common fields
	AliasID        string    `json:"alias_id"`
	OwnerDomainID  string    `json:"owner_domain_id"`
	TargetDomainID string    `json:"target_domain_id"`
	AliasType      string    `json:"alias_type"` // AUTO, RELATIONSHIP, MASK
	AliasName      string    `json:"alias_name"`
	CreatedAt      time.Time `json:"created_at,omitempty"`

	// Optional fields (for add receipts)
	IsCorporealAuto bool                   `json:"is_corporeal_auto,omitempty"`
	DSCIContractID  string                 `json:"dsci_contract_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`

	// Retirement fields (for remove receipts)
	RetiredAt time.Time `json:"retired_at,omitempty"`
	Reason    string    `json:"reason,omitempty"`
}

// GOV-12 Receipt Type Documentation
//
// identity.alias.add.v1:
//   Emitted when any alias (AUTO/RELATIONSHIP/MASK) is created.
//   Payload includes alias_id, owner_domain_id, target_domain_id, alias_type, alias_name.
//   For AUTO aliases: includes is_corporeal_auto flag.
//   For RELATIONSHIP aliases: includes dsci_contract_id (when available).
//   For MASK aliases: metadata may include ttl_days.
//
// identity.alias.remove.v1:
//   Emitted when an alias is retired (retired_at set).
//   Payload includes alias_id, retired_at, reason.
//   Reason examples: "user request", "TTL expired", "contract terminated", "domain exit".
//
// Receipt Continuity:
//   - All alias receipts chain via prev_id to establish provenance.
//   - Each receipt is SHA-256 hashed with payload + prev_id.
//   - Receipts are immutable and stored in identity_receipts table (GOV-10).
//   - Consent tracking: consent_by field links to the seat providing consent.
//
// Future Enhancements (GOV-13+):
//   - Link receipts to DSCI contracts table via dsci_contract_id FK.
//   - Add contract.terminate receipt that automatically triggers alias.remove for dependent aliases.
//   - Implement REGO policies that query receipt chain for authorization decisions.
