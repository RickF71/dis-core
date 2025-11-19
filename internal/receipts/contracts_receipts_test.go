package receipts

import (
	"context"
	"testing"
	"time"

	"dis-core/internal/contracts"
	"dis-core/internal/identity"

	"github.com/google/uuid"
)

// GOV-13: Contract Receipt Tests

// MockContractReceiptRecorder captures receipt calls for testing
type MockContractReceiptRecorder struct {
	RecordedReceipts []*identity.IdentityReceipt
	ShouldFail       bool
}

func (m *MockContractReceiptRecorder) RecordIdentityReceipt(ctx context.Context, receipt *identity.IdentityReceipt) error {
	if m.ShouldFail {
		return nil // Return error for failure case
	}
	m.RecordedReceipts = append(m.RecordedReceipts, receipt)
	return nil
}

func TestRecordContractCreateReceipt(t *testing.T) {
	ctx := context.Background()
	mockStore := &MockContractReceiptRecorder{}

	issuerID := uuid.New()
	subjectID := uuid.New()
	aliasID := uuid.New()

	// Create test contract
	contract := &contracts.Contract{
		ID:              uuid.New(),
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		AliasID:         &aliasID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership/v1",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
		ExpiresAt:       nil,
		Status:          contracts.ContractStatusActive,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	// Record receipt
	err := RecordContractCreateReceipt(ctx, mockStore, contract)
	if err != nil {
		t.Fatalf("RecordContractCreateReceipt failed: %v", err)
	}

	// Assert receipt recorded
	if len(mockStore.RecordedReceipts) != 1 {
		t.Fatalf("Expected 1 receipt, got %d", len(mockStore.RecordedReceipts))
	}

	receipt := mockStore.RecordedReceipts[0]

	// Assert receipt fields
	if receipt.Action != identity.ContractCreateV1 {
		t.Errorf("Expected action contract.create.v1, got %s", receipt.Action)
	}

	if receipt.DomainID != issuerID {
		t.Errorf("Expected domain_id %s, got %s", issuerID, receipt.DomainID)
	}

	if receipt.ActorID != issuerID {
		t.Errorf("Expected actor_id %s, got %s", issuerID, receipt.ActorID)
	}

	if receipt.ConsentBy != subjectID {
		t.Errorf("Expected consent_by %s, got %s", subjectID, receipt.ConsentBy)
	}

	// Assert payload contains required fields
	if receipt.Payload == nil {
		t.Fatal("Expected payload to be set")
	}

	contractID, ok := receipt.Payload["contract_id"].(string)
	if !ok || contractID != contract.ID.String() {
		t.Errorf("Expected contract_id %s in payload, got %v", contract.ID, receipt.Payload["contract_id"])
	}

	domainID, ok := receipt.Payload["domain_id"].(string)
	if !ok || domainID != issuerID.String() {
		t.Errorf("Expected domain_id %s in payload, got %v", issuerID, receipt.Payload["domain_id"])
	}

	subjectDomainID, ok := receipt.Payload["subject_domain_id"].(string)
	if !ok || subjectDomainID != subjectID.String() {
		t.Errorf("Expected subject_domain_id %s in payload, got %v", subjectID, receipt.Payload["subject_domain_id"])
	}

	contractType, ok := receipt.Payload["contract_type"].(string)
	if !ok || contractType != "membership" {
		t.Errorf("Expected contract_type 'membership' in payload, got %v", receipt.Payload["contract_type"])
	}

	dsciChannel, ok := receipt.Payload["dsci_channel"].(string)
	if !ok || dsciChannel != "web" {
		t.Errorf("Expected dsci_channel 'web' in payload, got %v", receipt.Payload["dsci_channel"])
	}

	dsciReference, ok := receipt.Payload["dsci_reference"].(string)
	if !ok || dsciReference != "https://example.com/membership/v1" {
		t.Errorf("Expected dsci_reference in payload, got %v", receipt.Payload["dsci_reference"])
	}

	dsciVersion, ok := receipt.Payload["dsci_version"].(string)
	if !ok || dsciVersion != "1.0.0" {
		t.Errorf("Expected dsci_version '1.0.0' in payload, got %v", receipt.Payload["dsci_version"])
	}

	status, ok := receipt.Payload["status"].(string)
	if !ok || status != string(contracts.ContractStatusActive) {
		t.Errorf("Expected status 'active' in payload, got %v", receipt.Payload["status"])
	}

	// Assert optional alias_id present
	aliasIDStr, ok := receipt.Payload["alias_id"].(string)
	if !ok || aliasIDStr != aliasID.String() {
		t.Errorf("Expected alias_id %s in payload, got %v", aliasID, receipt.Payload["alias_id"])
	}
}

func TestRecordContractCreateReceipt_NoAliasID(t *testing.T) {
	ctx := context.Background()
	mockStore := &MockContractReceiptRecorder{}

	issuerID := uuid.New()
	subjectID := uuid.New()

	// Create contract without alias_id
	contract := &contracts.Contract{
		ID:              uuid.New(),
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		AliasID:         nil, // No alias
		ContractType:    "tos",
		DSCIChannel:     "api",
		DSCIReference:   "https://api.example.com/tos",
		DSCIVersion:     "2.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
		Status:          contracts.ContractStatusActive,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	// Record receipt
	err := RecordContractCreateReceipt(ctx, mockStore, contract)
	if err != nil {
		t.Fatalf("RecordContractCreateReceipt failed: %v", err)
	}

	// Assert receipt recorded
	if len(mockStore.RecordedReceipts) != 1 {
		t.Fatalf("Expected 1 receipt, got %d", len(mockStore.RecordedReceipts))
	}

	receipt := mockStore.RecordedReceipts[0]

	// Assert alias_id not in payload when nil
	if _, exists := receipt.Payload["alias_id"]; exists {
		t.Error("Expected alias_id to not be in payload when nil")
	}
}

func TestRecordContractCreateReceipt_WithExpiry(t *testing.T) {
	ctx := context.Background()
	mockStore := &MockContractReceiptRecorder{}

	issuerID := uuid.New()
	subjectID := uuid.New()
	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour) // 30 days

	// Create contract with expires_at
	contract := &contracts.Contract{
		ID:              uuid.New(),
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "subscription",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/subscription",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
		ExpiresAt:       &expiresAt,
		Status:          contracts.ContractStatusActive,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	// Record receipt
	err := RecordContractCreateReceipt(ctx, mockStore, contract)
	if err != nil {
		t.Fatalf("RecordContractCreateReceipt failed: %v", err)
	}

	receipt := mockStore.RecordedReceipts[0]

	// Assert expires_at in payload
	expiresAtStr, ok := receipt.Payload["expires_at"].(string)
	if !ok {
		t.Error("Expected expires_at to be in payload")
	} else {
		parsedExpiry, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			t.Errorf("Failed to parse expires_at: %v", err)
		}
		if parsedExpiry.Sub(expiresAt).Abs() > time.Second {
			t.Errorf("Expected expires_at near %s, got %s", expiresAt, parsedExpiry)
		}
	}
}

func TestRecordContractRevokeReceipt(t *testing.T) {
	ctx := context.Background()
	mockStore := &MockContractReceiptRecorder{}

	issuerID := uuid.New()
	subjectID := uuid.New()
	revokedAt := time.Now().UTC()

	// Create revoked contract
	contract := &contracts.Contract{
		ID:              uuid.New(),
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-30 * 24 * time.Hour), // 30 days ago
		RevokedAt:       &revokedAt,
		Status:          contracts.ContractStatusRevoked,
		CreatedAt:       time.Now().UTC().Add(-30 * 24 * time.Hour),
		UpdatedAt:       time.Now().UTC(),
	}

	// Record revoke receipt
	reason := "User requested cancellation"
	actorID := issuerID.String()
	err := RecordContractRevokeReceipt(ctx, mockStore, contract, reason, actorID)
	if err != nil {
		t.Fatalf("RecordContractRevokeReceipt failed: %v", err)
	}

	// Assert receipt recorded
	if len(mockStore.RecordedReceipts) != 1 {
		t.Fatalf("Expected 1 receipt, got %d", len(mockStore.RecordedReceipts))
	}

	receipt := mockStore.RecordedReceipts[0]

	// Assert receipt fields
	if receipt.Action != identity.ContractRevokeV1 {
		t.Errorf("Expected action contract.revoke.v1, got %s", receipt.Action)
	}

	if receipt.DomainID != issuerID {
		t.Errorf("Expected domain_id %s, got %s", issuerID, receipt.DomainID)
	}

	if receipt.ActorID != issuerID {
		t.Errorf("Expected actor_id %s, got %s", issuerID, receipt.ActorID)
	}

	if receipt.ConsentBy != subjectID {
		t.Errorf("Expected consent_by %s, got %s", subjectID, receipt.ConsentBy)
	}

	// Assert payload contains required fields
	if receipt.Payload == nil {
		t.Fatal("Expected payload to be set")
	}

	contractID, ok := receipt.Payload["contract_id"].(string)
	if !ok || contractID != contract.ID.String() {
		t.Errorf("Expected contract_id %s in payload, got %v", contract.ID, receipt.Payload["contract_id"])
	}

	revokedAtStr, ok := receipt.Payload["revoked_at"].(string)
	if !ok {
		t.Error("Expected revoked_at in payload")
	} else {
		parsedRevoked, err := time.Parse(time.RFC3339, revokedAtStr)
		if err != nil {
			t.Errorf("Failed to parse revoked_at: %v", err)
		}
		if parsedRevoked.Sub(revokedAt).Abs() > time.Second {
			t.Errorf("Expected revoked_at near %s, got %s", revokedAt, parsedRevoked)
		}
	}

	revocationReason, ok := receipt.Payload["revocation_reason"].(string)
	if !ok || revocationReason != reason {
		t.Errorf("Expected revocation_reason '%s' in payload, got %v", reason, receipt.Payload["revocation_reason"])
	}

	previousStatus, ok := receipt.Payload["previous_status"].(string)
	if !ok || previousStatus != "active" {
		t.Errorf("Expected previous_status 'active' in payload, got %v", receipt.Payload["previous_status"])
	}

	// Assert DSCI triple in payload
	dsciChannel, ok := receipt.Payload["dsci_channel"].(string)
	if !ok || dsciChannel != "web" {
		t.Errorf("Expected dsci_channel 'web' in payload, got %v", receipt.Payload["dsci_channel"])
	}

	dsciReference, ok := receipt.Payload["dsci_reference"].(string)
	if !ok || dsciReference != "https://example.com/membership" {
		t.Errorf("Expected dsci_reference in payload, got %v", receipt.Payload["dsci_reference"])
	}

	dsciVersion, ok := receipt.Payload["dsci_version"].(string)
	if !ok || dsciVersion != "1.0.0" {
		t.Errorf("Expected dsci_version '1.0.0' in payload, got %v", receipt.Payload["dsci_version"])
	}
}

func TestRecordContractRevokeReceipt_NilStore(t *testing.T) {
	ctx := context.Background()

	issuerID := uuid.New()
	subjectID := uuid.New()
	revokedAt := time.Now().UTC()

	contract := &contracts.Contract{
		ID:              uuid.New(),
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "tos",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/tos",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
		RevokedAt:       &revokedAt,
		Status:          contracts.ContractStatusRevoked,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	// Call with nil store (graceful degradation)
	err := RecordContractRevokeReceipt(ctx, nil, contract, "test", issuerID.String())
	if err != nil {
		t.Errorf("Expected nil error for nil store, got: %v", err)
	}
}

func TestRecordContractRevokeReceipt_MissingRevokedAt(t *testing.T) {
	ctx := context.Background()
	mockStore := &MockContractReceiptRecorder{}

	issuerID := uuid.New()
	subjectID := uuid.New()

	// Contract without revoked_at (invalid state)
	contract := &contracts.Contract{
		ID:              uuid.New(),
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
		RevokedAt:       nil, // Missing
		Status:          contracts.ContractStatusRevoked,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	// Should return error
	err := RecordContractRevokeReceipt(ctx, mockStore, contract, "test", issuerID.String())
	if err == nil {
		t.Error("Expected error when revoked_at is nil, got nil")
	}
}
