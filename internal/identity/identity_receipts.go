package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-10: Identity Provenance & Alias Receipt Integration
// GOV-11: Domain-Scoped Identity Projection & Corporeal Authentication
// Extends GOV-9 receipt ledger to cover identity binding, alias management,
// notech→dis conversion, corporeal-root authority provenance, domain projections,
// and foreign identity acceptance.

// IdentityAction represents types of identity governance actions
type IdentityAction string

const (
	// GOV-10 actions
	IdentityRootV1          IdentityAction = "identity.root.v1"           // Root identity establishment
	IdentityAliasAddV1      IdentityAction = "identity.alias.add.v1"      // Alias addition
	IdentityAliasRemoveV1   IdentityAction = "identity.alias.remove.v1"   // Alias removal
	IdentityConvertV1       IdentityAction = "identity.convert.v1"        // notech→dis conversion
	IdentityBindingUpdateV1 IdentityAction = "identity.binding.update.v1" // Binding update

	// GOV-11 actions: Domain-scoped identity projections
	IdentityDomainIDCreateV1 IdentityAction = "identity.domainid.create.v1" // Domain creates local identity projection
	IdentityDomainIDUpdateV1 IdentityAction = "identity.domainid.update.v1" // Domain updates/rotates local identity
	IdentityAcceptV1         IdentityAction = "identity.accept.v1"          // Domain accepts foreign identity
	IdentityAcceptRevokeV1   IdentityAction = "identity.accept.revoke.v1"   // Domain revokes foreign identity acceptance
	IdentityIRLAuthV1        IdentityAction = "identity.irlauth.v1"         // Corporeal domain authenticates IRL

	// GOV-13 actions: DSCI contract operations
	ContractCreateV1 IdentityAction = "contract.create.v1" // Contract creation
	ContractRevokeV1 IdentityAction = "contract.revoke.v1" // Contract revocation
)

// IdentityReceipt represents an immutable identity governance action
type IdentityReceipt struct {
	ID         uuid.UUID              `json:"id"`
	DomainID   uuid.UUID              `json:"domain_id"`
	ActorID    uuid.UUID              `json:"actor_id"`
	Action     IdentityAction         `json:"action"`
	Payload    map[string]interface{} `json:"payload"`
	PrevID     *uuid.UUID             `json:"prev_id,omitempty"`
	Hash       string                 `json:"hash"`
	CreatedAt  time.Time              `json:"created_at"`
	ConsentBy  uuid.UUID              `json:"consent_by"`
	AliasScope *string                `json:"alias_scope,omitempty"`

	// GOV-11: Domain projection and foreign identity fields
	TargetDomainID  *uuid.UUID `json:"target_domain_id,omitempty"` // Target domain for domain.id operations
	SourceDomainID  *uuid.UUID `json:"source_domain_id,omitempty"` // Source domain for foreign identity acceptance
	ExternalSubject *string    `json:"external_subject,omitempty"` // External identity subject/token
	Channel         *string    `json:"channel,omitempty"`          // Authentication channel (IRL auth)
	Method          *string    `json:"method,omitempty"`           // Authentication method (IRL auth)
	Scope           *string    `json:"scope,omitempty"`            // Foreign identity acceptance scope
}

// IdentityReceiptStore manages identity receipt persistence
type IdentityReceiptStore struct {
	db *pgxpool.Pool
}

// NewIdentityReceiptStore creates a new identity receipt store
func NewIdentityReceiptStore(db *pgxpool.Pool) *IdentityReceiptStore {
	return &IdentityReceiptStore{db: db}
}

// computeIdentityHash computes SHA-256 hash of payload + prev_id
// GOV-10: Similar to GOV-9 authority receipts but for identity operations
func computeIdentityHash(payload []byte, prevID *uuid.UUID) string {
	h := sha256.New()
	h.Write(payload)
	if prevID != nil {
		h.Write([]byte(prevID.String()))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// RecordIdentityReceipt creates an immutable identity receipt with hash chain
func (s *IdentityReceiptStore) RecordIdentityReceipt(ctx context.Context, r *IdentityReceipt) error {
	// Marshal payload to bytes for hash computation
	payloadBytes, err := json.Marshal(r.Payload)
	if err != nil {
		return err
	}

	// Compute hash: SHA-256(payload + prev_id)
	r.Hash = computeIdentityHash(payloadBytes, r.PrevID)

	// Set creation timestamp if not provided
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}

	// Generate ID if not provided
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}

	// Insert immutable receipt
	_, err = s.db.Exec(ctx, `
		INSERT INTO identity_receipts
		  (id, domain_id, actor_id, action, payload, prev_id, hash, created_at, consent_by, alias_scope,
		   target_domain_id, source_domain_id, external_subject, channel, method, scope)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`,
		r.ID,
		r.DomainID,
		r.ActorID,
		string(r.Action),
		payloadBytes,
		r.PrevID,
		r.Hash,
		r.CreatedAt,
		r.ConsentBy,
		r.AliasScope,
		r.TargetDomainID,
		r.SourceDomainID,
		r.ExternalSubject,
		r.Channel,
		r.Method,
		r.Scope,
	)

	return err
}

// GetLastReceipt retrieves the most recent identity receipt for an actor
func (s *IdentityReceiptStore) GetLastReceipt(ctx context.Context, actorID uuid.UUID) (*IdentityReceipt, error) {
	var receipt IdentityReceipt
	var payloadBytes []byte
	var action string

	err := s.db.QueryRow(ctx, `
		SELECT id, domain_id, actor_id, action, payload, prev_id, hash, created_at, consent_by, alias_scope,
		       target_domain_id, source_domain_id, external_subject, channel, method, scope
		FROM identity_receipts
		WHERE actor_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, actorID).Scan(
		&receipt.ID,
		&receipt.DomainID,
		&receipt.ActorID,
		&action,
		&payloadBytes,
		&receipt.PrevID,
		&receipt.Hash,
		&receipt.CreatedAt,
		&receipt.ConsentBy,
		&receipt.AliasScope,
		&receipt.TargetDomainID,
		&receipt.SourceDomainID,
		&receipt.ExternalSubject,
		&receipt.Channel,
		&receipt.Method,
		&receipt.Scope,
	)

	if err != nil {
		return nil, err
	}

	receipt.Action = IdentityAction(action)

	// Unmarshal payload
	if err := json.Unmarshal(payloadBytes, &receipt.Payload); err != nil {
		return nil, err
	}

	return &receipt, nil
}

// RecordIdentityAction is a convenience wrapper for common identity actions
func (s *IdentityReceiptStore) RecordIdentityAction(
	ctx context.Context,
	actorID uuid.UUID,
	domainID uuid.UUID,
	action IdentityAction,
	payload map[string]interface{},
	consentBy uuid.UUID,
	aliasScope *string,
) (*IdentityReceipt, error) {
	// Get previous receipt to chain
	prevReceipt, err := s.GetLastReceipt(ctx, actorID)
	var prevID *uuid.UUID
	if err == nil && prevReceipt != nil {
		prevID = &prevReceipt.ID
	}

	receipt := &IdentityReceipt{
		ID:         uuid.New(),
		DomainID:   domainID,
		ActorID:    actorID,
		Action:     action,
		Payload:    payload,
		PrevID:     prevID,
		ConsentBy:  consentBy,
		AliasScope: aliasScope,
	}

	if err := s.RecordIdentityReceipt(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

// GOV-11: Helper functions for domain-scoped identity operations

// RecordDomainIDCreation records when a domain creates a local identity projection for an actor
func (s *IdentityReceiptStore) RecordDomainIDCreation(
	ctx context.Context,
	actorID uuid.UUID,
	domainID uuid.UUID,
	domainIdentityValue string,
	context map[string]interface{},
) (*IdentityReceipt, error) {
	payload := map[string]interface{}{
		"domain_identity_value": domainIdentityValue,
		"action_type":           "create",
	}
	if context != nil {
		payload["context"] = context
	}

	prevReceipt, _ := s.GetLastReceipt(ctx, actorID)
	var prevID *uuid.UUID
	if prevReceipt != nil {
		prevID = &prevReceipt.ID
	}

	receipt := &IdentityReceipt{
		ID:             uuid.New(),
		DomainID:       domainID,
		ActorID:        actorID,
		Action:         IdentityDomainIDCreateV1,
		Payload:        payload,
		PrevID:         prevID,
		ConsentBy:      domainID,
		TargetDomainID: &domainID,
	}

	if err := s.RecordIdentityReceipt(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

// RecordForeignIdentityAcceptance records when a domain accepts a foreign identity for an actor
func (s *IdentityReceiptStore) RecordForeignIdentityAcceptance(
	ctx context.Context,
	actorID uuid.UUID,
	acceptingDomainID uuid.UUID,
	sourceDomainID uuid.UUID,
	externalSubject string,
	scope *string,
	context map[string]interface{},
) (*IdentityReceipt, error) {
	payload := map[string]interface{}{
		"accepting_domain": acceptingDomainID.String(),
		"source_domain":    sourceDomainID.String(),
		"external_subject": externalSubject,
	}
	if scope != nil {
		payload["scope"] = *scope
	}
	if context != nil {
		payload["context"] = context
	}

	prevReceipt, _ := s.GetLastReceipt(ctx, actorID)
	var prevID *uuid.UUID
	if prevReceipt != nil {
		prevID = &prevReceipt.ID
	}

	receipt := &IdentityReceipt{
		ID:              uuid.New(),
		DomainID:        acceptingDomainID,
		ActorID:         actorID,
		Action:          IdentityAcceptV1,
		Payload:         payload,
		PrevID:          prevID,
		ConsentBy:       acceptingDomainID,
		TargetDomainID:  &acceptingDomainID,
		SourceDomainID:  &sourceDomainID,
		ExternalSubject: &externalSubject,
		Scope:           scope,
	}

	if err := s.RecordIdentityReceipt(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

// RecordIRLAuthentication records when a corporeal domain authenticates its person for IRL context
func (s *IdentityReceiptStore) RecordIRLAuthentication(
	ctx context.Context,
	actorID uuid.UUID,
	corporealDomainID uuid.UUID,
	targetDomainID uuid.UUID,
	method string,
	channel string,
	context map[string]interface{},
) (*IdentityReceipt, error) {
	payload := map[string]interface{}{
		"corporeal_domain": corporealDomainID.String(),
		"target_domain":    targetDomainID.String(),
		"method":           method,
		"channel":          channel,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}
	if context != nil {
		payload["context"] = context
	}

	prevReceipt, _ := s.GetLastReceipt(ctx, actorID)
	var prevID *uuid.UUID
	if prevReceipt != nil {
		prevID = &prevReceipt.ID
	}

	receipt := &IdentityReceipt{
		ID:             uuid.New(),
		DomainID:       corporealDomainID,
		ActorID:        actorID,
		Action:         IdentityIRLAuthV1,
		Payload:        payload,
		PrevID:         prevID,
		ConsentBy:      corporealDomainID,
		TargetDomainID: &targetDomainID,
		SourceDomainID: &corporealDomainID,
		Channel:        &channel,
		Method:         &method,
	}

	if err := s.RecordIdentityReceipt(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

// RecordForeignIdentityRevocation records when a domain revokes a previously accepted foreign identity
func (s *IdentityReceiptStore) RecordForeignIdentityRevocation(
	ctx context.Context,
	actorID uuid.UUID,
	revokingDomainID uuid.UUID,
	sourceDomainID uuid.UUID,
	externalSubject string,
	reason string,
) (*IdentityReceipt, error) {
	payload := map[string]interface{}{
		"revoking_domain":  revokingDomainID.String(),
		"source_domain":    sourceDomainID.String(),
		"external_subject": externalSubject,
		"reason":           reason,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	prevReceipt, _ := s.GetLastReceipt(ctx, actorID)
	var prevID *uuid.UUID
	if prevReceipt != nil {
		prevID = &prevReceipt.ID
	}

	receipt := &IdentityReceipt{
		ID:              uuid.New(),
		DomainID:        revokingDomainID,
		ActorID:         actorID,
		Action:          IdentityAcceptRevokeV1,
		Payload:         payload,
		PrevID:          prevID,
		ConsentBy:       revokingDomainID,
		TargetDomainID:  &revokingDomainID,
		SourceDomainID:  &sourceDomainID,
		ExternalSubject: &externalSubject,
	}

	if err := s.RecordIdentityReceipt(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

// GOV-11G: Schema-Aware Identity Policy Editing
const (
	IdentitySchemaAdoptV1  IdentityAction = "identity.schema.adopt.v1"  // Domain adopts an identity schema
	IdentityPolicyUpdateV1 IdentityAction = "identity.policy.update.v1" // Domain updates identity policy
)

// RecordSchemaAdoption records when a domain adopts an identity schema
func (s *IdentityReceiptStore) RecordSchemaAdoption(
	ctx context.Context,
	domainID uuid.UUID,
	schemaID string,
	schemaVersion string,
	actorID *uuid.UUID,
	metadata map[string]interface{},
) (*IdentityReceipt, error) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["schema_id"] = schemaID
	metadata["schema_version"] = schemaVersion
	metadata["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	var actor uuid.UUID
	if actorID != nil {
		actor = *actorID
	} else {
		actor = domainID // Default to domain if no actor specified
	}

	prevReceipt, _ := s.GetLastReceipt(ctx, actor)
	var prevID *uuid.UUID
	if prevReceipt != nil {
		prevID = &prevReceipt.ID
	}

	receipt := &IdentityReceipt{
		ID:         uuid.New(),
		DomainID:   domainID,
		ActorID:    actor,
		Action:     IdentitySchemaAdoptV1,
		Payload:    metadata,
		PrevID:     prevID,
		ConsentBy:  domainID,
		AliasScope: &schemaID, // Reuse alias_scope for schema tracking
	}

	if err := s.RecordIdentityReceipt(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

// RecordPolicyUpdate records when a domain updates its identity policy
func (s *IdentityReceiptStore) RecordPolicyUpdate(
	ctx context.Context,
	domainID uuid.UUID,
	actorID *uuid.UUID,
	metadata map[string]interface{},
) (*IdentityReceipt, error) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	var actor uuid.UUID
	if actorID != nil {
		actor = *actorID
	} else {
		actor = domainID
	}

	prevReceipt, _ := s.GetLastReceipt(ctx, actor)
	var prevID *uuid.UUID
	if prevReceipt != nil {
		prevID = &prevReceipt.ID
	}

	receipt := &IdentityReceipt{
		ID:        uuid.New(),
		DomainID:  domainID,
		ActorID:   actor,
		Action:    IdentityPolicyUpdateV1,
		Payload:   metadata,
		PrevID:    prevID,
		ConsentBy: domainID,
	}

	if err := s.RecordIdentityReceipt(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}
