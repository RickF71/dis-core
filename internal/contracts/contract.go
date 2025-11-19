package contracts

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GOV-13: Contracts Table & DSCI Contract Wiring

// ContractStatus represents the lifecycle state of a contract
type ContractStatus string

const (
	ContractStatusPending ContractStatus = "pending" // Awaiting activation
	ContractStatusActive  ContractStatus = "active"  // Currently in effect
	ContractStatusExpired ContractStatus = "expired" // Past expiration date
	ContractStatusRevoked ContractStatus = "revoked" // Explicitly revoked
)

// IsValid checks if the contract status is one of the defined values
func (s ContractStatus) IsValid() bool {
	switch s {
	case ContractStatusPending, ContractStatusActive, ContractStatusExpired, ContractStatusRevoked:
		return true
	default:
		return false
	}
}

// Contract represents a DSCI (Domain-Signed Contract Inheritance) contract
// that governs domain participation, membership, data processing, and other consent relationships.
type Contract struct {
	ID              uuid.UUID      `json:"id"`
	DomainID        uuid.UUID      `json:"domain_id"`         // Issuing domain (contract owner)
	SubjectDomainID uuid.UUID      `json:"subject_domain_id"` // Domain subject to contract (usually RELATIONSHIP alias)
	AliasID         *uuid.UUID     `json:"alias_id,omitempty"`
	ContractType    string         `json:"contract_type"`
	DSCIChannel     string         `json:"dsci_channel"`
	DSCIReference   string         `json:"dsci_reference"`
	DSCIVersion     string         `json:"dsci_version"`
	EffectiveAt     time.Time      `json:"effective_at"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	RevokedAt       *time.Time     `json:"revoked_at,omitempty"`
	Status          ContractStatus `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CreatedBy       *uuid.UUID     `json:"created_by,omitempty"`
}

// IsActive returns true if the contract is currently active (not pending, expired, or revoked)
func (c *Contract) IsActive() bool {
	if c.Status != ContractStatusActive {
		return false
	}
	if c.RevokedAt != nil {
		return false
	}
	now := time.Now()
	if c.ExpiresAt != nil && c.ExpiresAt.Before(now) {
		return false
	}
	return true
}

// IsExpired returns true if the contract has expired (past expires_at timestamp)
func (c *Contract) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsRevoked returns true if the contract has been revoked
func (c *Contract) IsRevoked() bool {
	return c.Status == ContractStatusRevoked || c.RevokedAt != nil
}

// Validate checks that the contract has all required fields and valid relationships
func (c *Contract) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("contract ID is required")
	}
	if c.DomainID == uuid.Nil {
		return fmt.Errorf("domain_id is required")
	}
	if c.SubjectDomainID == uuid.Nil {
		return fmt.Errorf("subject_domain_id is required")
	}
	if c.ContractType == "" {
		return fmt.Errorf("contract_type is required")
	}
	if c.DSCIChannel == "" {
		return fmt.Errorf("dsci_channel is required")
	}
	if c.DSCIReference == "" {
		return fmt.Errorf("dsci_reference is required")
	}
	if c.DSCIVersion == "" {
		return fmt.Errorf("dsci_version is required")
	}
	if c.EffectiveAt.IsZero() {
		return fmt.Errorf("effective_at is required")
	}
	if !c.Status.IsValid() {
		return fmt.Errorf("invalid contract status: %s", c.Status)
	}
	if c.ExpiresAt != nil && c.ExpiresAt.Before(c.EffectiveAt) {
		return fmt.Errorf("expires_at must be after effective_at")
	}
	if c.Status == ContractStatusRevoked && c.RevokedAt == nil {
		return fmt.Errorf("revoked contracts must have revoked_at timestamp")
	}
	if c.Status != ContractStatusRevoked && c.RevokedAt != nil {
		return fmt.Errorf("revoked_at can only be set for revoked contracts")
	}
	return nil
}

// ComputeStatus determines the appropriate status based on timestamps
// This is used during creation and updates to ensure status consistency
func (c *Contract) ComputeStatus() ContractStatus {
	if c.RevokedAt != nil {
		return ContractStatusRevoked
	}

	now := time.Now()

	if c.ExpiresAt != nil && c.ExpiresAt.Before(now) {
		return ContractStatusExpired
	}

	if c.EffectiveAt.After(now) {
		return ContractStatusPending
	}

	return ContractStatusActive
}

// GetDSCITriple returns the canonical DSCI triple (channel, reference, version)
func (c *Contract) GetDSCITriple() (channel, reference, version string) {
	return c.DSCIChannel, c.DSCIReference, c.DSCIVersion
}

// CreateContractInput represents input for creating a new contract
type CreateContractInput struct {
	DomainID        uuid.UUID  `json:"domain_id"`
	SubjectDomainID uuid.UUID  `json:"subject_domain_id"`
	AliasID         *uuid.UUID `json:"alias_id,omitempty"`
	ContractType    string     `json:"contract_type"`
	DSCIChannel     string     `json:"dsci_channel"`
	DSCIReference   string     `json:"dsci_reference"`
	DSCIVersion     string     `json:"dsci_version"`
	EffectiveAt     time.Time  `json:"effective_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty"`
}

// Validate checks that the input has all required fields
func (i *CreateContractInput) Validate() error {
	if i.DomainID == uuid.Nil {
		return fmt.Errorf("domain_id is required")
	}
	if i.SubjectDomainID == uuid.Nil {
		return fmt.Errorf("subject_domain_id is required")
	}
	if i.ContractType == "" {
		return fmt.Errorf("contract_type is required")
	}
	if i.DSCIChannel == "" {
		return fmt.Errorf("dsci_channel is required")
	}
	if i.DSCIReference == "" {
		return fmt.Errorf("dsci_reference is required")
	}
	if i.DSCIVersion == "" {
		return fmt.Errorf("dsci_version is required")
	}
	if i.EffectiveAt.IsZero() {
		return fmt.Errorf("effective_at is required")
	}
	if i.ExpiresAt != nil && i.ExpiresAt.Before(i.EffectiveAt) {
		return fmt.Errorf("expires_at must be after effective_at")
	}
	return nil
}

// ToContract converts CreateContractInput to a Contract with computed status
func (i *CreateContractInput) ToContract() *Contract {
	contract := &Contract{
		ID:              uuid.New(),
		DomainID:        i.DomainID,
		SubjectDomainID: i.SubjectDomainID,
		AliasID:         i.AliasID,
		ContractType:    i.ContractType,
		DSCIChannel:     i.DSCIChannel,
		DSCIReference:   i.DSCIReference,
		DSCIVersion:     i.DSCIVersion,
		EffectiveAt:     i.EffectiveAt,
		ExpiresAt:       i.ExpiresAt,
		CreatedBy:       i.CreatedBy,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Compute initial status
	contract.Status = contract.ComputeStatus()

	return contract
}

// ContractListResponse represents a list of contracts with metadata
type ContractListResponse struct {
	Contracts []Contract `json:"contracts"`
	Count     int        `json:"count"`
}
