package receipts

import (
	"context"
	"fmt"
	"time"

	"dis-core/internal/contracts"
	"dis-core/internal/identity"

	"github.com/google/uuid"
)

// GOV-13: Contract Receipts - Receipt Helpers
// These functions provide convenience wrappers for recording contract-related receipts
// using the existing GOV-10 identity receipt system.

// ContractReceiptPayload represents the canonical payload structure for contract receipts.
// This serves as documentation and is not instantiated directly.
type ContractReceiptPayload struct {
	// contract.create.v1 fields
	ContractID      string `json:"contract_id"`
	DomainID        string `json:"domain_id"`
	SubjectDomainID string `json:"subject_domain_id"`
	AliasID         string `json:"alias_id,omitempty"`
	ContractType    string `json:"contract_type"`
	DSCIChannel     string `json:"dsci_channel"`
	DSCIReference   string `json:"dsci_reference"`
	DSCIVersion     string `json:"dsci_version"`
	EffectiveAt     string `json:"effective_at"`
	ExpiresAt       string `json:"expires_at,omitempty"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`

	// contract.revoke.v1 fields
	RevokedAt        string `json:"revoked_at,omitempty"`
	RevocationReason string `json:"revocation_reason,omitempty"`
}

// RecordContractCreateReceipt records a contract.create.v1 receipt
// for DSCI contract creation.
//
// Parameters:
// - ctx: request context
// - store: identity receipt store (GOV-10)
// - contract: the newly created contract
//
// Receipt Type: contract.create.v1
// Purpose: Immutable record of contract creation with DSCI triple
func RecordContractCreateReceipt(
	ctx context.Context,
	store ReceiptRecorder,
	contract *contracts.Contract,
) error {
	if store == nil {
		return nil // Graceful degradation if receipt system not configured
	}

	// Build payload with contract metadata and DSCI triple
	payload := map[string]interface{}{
		"contract_id":       contract.ID.String(),
		"domain_id":         contract.DomainID.String(),
		"subject_domain_id": contract.SubjectDomainID.String(),
		"contract_type":     contract.ContractType,
		"dsci_channel":      contract.DSCIChannel,
		"dsci_reference":    contract.DSCIReference,
		"dsci_version":      contract.DSCIVersion,
		"effective_at":      contract.EffectiveAt.Format(time.RFC3339),
		"status":            string(contract.Status),
		"created_at":        contract.CreatedAt.Format(time.RFC3339),
	}

	// Add optional fields
	if contract.AliasID != nil {
		payload["alias_id"] = contract.AliasID.String()
	}
	if contract.ExpiresAt != nil {
		payload["expires_at"] = contract.ExpiresAt.Format(time.RFC3339)
	}
	if contract.CreatedBy != nil {
		payload["created_by"] = contract.CreatedBy.String()
	}

	// Record receipt using identity receipt system
	receipt := &identity.IdentityReceipt{
		ID:        uuid.New(),
		DomainID:  contract.DomainID, // Domain issuing the contract
		ActorID:   contract.DomainID, // Domain issuing the contract
		Action:    identity.ContractCreateV1,
		ConsentBy: contract.SubjectDomainID, // Subject consents
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}

	if err := store.RecordIdentityReceipt(ctx, receipt); err != nil {
		return fmt.Errorf("failed to record contract create receipt: %w", err)
	}

	return nil
}

// RecordContractRevokeReceipt records a contract.revoke.v1 receipt
// for DSCI contract revocation.
//
// Parameters:
// - ctx: request context
// - store: identity receipt store (GOV-10)
// - contract: the revoked contract
// - reason: human-readable revocation reason
// - actorID: the actor performing the revocation (domain admin, system, etc.)
//
// Receipt Type: contract.revoke.v1
// Purpose: Immutable record of contract revocation for audit and continuity
func RecordContractRevokeReceipt(
	ctx context.Context,
	store ReceiptRecorder,
	contract *contracts.Contract,
	reason string,
	actorID string,
) error {
	if store == nil {
		return nil // Graceful degradation if receipt system not configured
	}

	if contract.RevokedAt == nil {
		return fmt.Errorf("contract.revoked_at is nil, cannot record revoke receipt")
	}

	// Parse actorID to UUID
	actorUUID, err := uuid.Parse(actorID)
	if err != nil {
		// Fallback to domain_id if actorID is invalid
		actorUUID = contract.DomainID
	}

	// Build payload with revocation metadata
	payload := map[string]interface{}{
		"contract_id":       contract.ID.String(),
		"revoked_at":        contract.RevokedAt.Format(time.RFC3339),
		"revocation_reason": reason,
		"previous_status":   "active", // Contracts can only be revoked from active state
		"domain_id":         contract.DomainID.String(),
		"subject_domain_id": contract.SubjectDomainID.String(),
		"contract_type":     contract.ContractType,
	}

	// Include DSCI triple for traceability
	payload["dsci_channel"] = contract.DSCIChannel
	payload["dsci_reference"] = contract.DSCIReference
	payload["dsci_version"] = contract.DSCIVersion

	// Record receipt using identity receipt system
	receipt := &identity.IdentityReceipt{
		ID:        uuid.New(),
		DomainID:  contract.DomainID, // Domain that issued the contract
		ActorID:   actorUUID,         // Actor performing revocation
		Action:    identity.ContractRevokeV1,
		ConsentBy: contract.SubjectDomainID, // Subject consents (implicit in revocation)
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}

	if err := store.RecordIdentityReceipt(ctx, receipt); err != nil {
		return fmt.Errorf("failed to record contract revoke receipt: %w", err)
	}

	return nil
}

// ReceiptRecorder is a minimal interface for recording receipts.
// This allows the contract service to depend on an interface rather than a concrete type.
type ReceiptRecorder interface {
	RecordIdentityReceipt(ctx context.Context, receipt *identity.IdentityReceipt) error
}

// Compile-time check that IdentityReceiptStore implements ReceiptRecorder
var _ ReceiptRecorder = (*identity.IdentityReceiptStore)(nil)
