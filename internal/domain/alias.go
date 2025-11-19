package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GOV-12: Alias Canon & DSCI Integration
// Canonical alias taxonomy: AUTO / RELATIONSHIP / MASK

// AliasType represents the canonical alias taxonomy
type AliasType string

const (
	// AliasTypeAuto represents structural, non-volitional graph bindings
	// Created automatically by system flows (corporeal Prime Seat, seat instantiation)
	AliasTypeAuto AliasType = "AUTO"

	// AliasTypeRelationship represents volitional, domain-scoped identity
	// Created when person explicitly joins domain with consent
	AliasTypeRelationship AliasType = "RELATIONSHIP"

	// AliasTypeMask represents ephemeral browsing personas
	// Created for exploration with minimal authority footprint
	AliasTypeMask AliasType = "MASK"
)

// String returns string representation of AliasType
func (at AliasType) String() string {
	return string(at)
}

// IsValid checks if the alias type is one of the canonical types
func (at AliasType) IsValid() bool {
	switch at {
	case AliasTypeAuto, AliasTypeRelationship, AliasTypeMask:
		return true
	default:
		return false
	}
}

// AliasStatus represents the lifecycle state of an alias
type AliasStatus string

const (
	// AliasStatusActive indicates alias is currently usable
	AliasStatusActive AliasStatus = "ACTIVE"

	// AliasStatusRetired indicates alias has been deactivated
	AliasStatusRetired AliasStatus = "RETIRED"
)

// String returns string representation of AliasStatus
func (as AliasStatus) String() string {
	return string(as)
}

// AliasMetadata represents JSONB metadata for aliases
type AliasMetadata map[string]interface{}

// Value implements driver.Valuer for database storage
func (am AliasMetadata) Value() (driver.Value, error) {
	if am == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(am)
}

// Scan implements sql.Scanner for database retrieval
func (am *AliasMetadata) Scan(value interface{}) error {
	if value == nil {
		*am = make(AliasMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan AliasMetadata: expected []byte, got %T", value)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return fmt.Errorf("failed to unmarshal AliasMetadata: %w", err)
	}

	*am = AliasMetadata(data)
	return nil
}

// Alias represents a canonical DIS alias
type Alias struct {
	ID              uuid.UUID     `db:"id" json:"id"`
	OwnerDomainID   uuid.UUID     `db:"owner_domain_id" json:"owner_domain_id"`
	TargetDomainID  uuid.UUID     `db:"target_domain_id" json:"target_domain_id"`
	AliasName       string        `db:"alias_name" json:"alias_name"`
	AliasType       AliasType     `db:"alias_type" json:"alias_type"`
	IsCorporealAuto bool          `db:"is_corporeal_auto" json:"is_corporeal_auto"`
	DSCIContractID  *uuid.UUID    `db:"dsci_contract_id" json:"dsci_contract_id,omitempty"`
	CreatedAt       time.Time     `db:"created_at" json:"created_at"`
	RetiredAt       *time.Time    `db:"retired_at" json:"retired_at,omitempty"`
	Metadata        AliasMetadata `db:"metadata" json:"metadata"`
}

// Status returns the current status of the alias
func (a *Alias) Status() AliasStatus {
	if a.RetiredAt != nil {
		return AliasStatusRetired
	}
	return AliasStatusActive
}

// IsActive returns true if the alias is currently active
func (a *Alias) IsActive() bool {
	return a.Status() == AliasStatusActive
}

// IsMask returns true if this is a MASK alias
func (a *Alias) IsMask() bool {
	return a.AliasType == AliasTypeMask
}

// IsAuto returns true if this is an AUTO alias
func (a *Alias) IsAuto() bool {
	return a.AliasType == AliasTypeAuto
}

// IsRelationship returns true if this is a RELATIONSHIP alias
func (a *Alias) IsRelationship() bool {
	return a.AliasType == AliasTypeRelationship
}

// Validate performs basic validation on the alias
func (a *Alias) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("alias ID is required")
	}
	if a.OwnerDomainID == uuid.Nil {
		return fmt.Errorf("owner_domain_id is required")
	}
	if a.TargetDomainID == uuid.Nil {
		return fmt.Errorf("target_domain_id is required")
	}
	if a.AliasName == "" {
		return fmt.Errorf("alias_name is required")
	}
	if !a.AliasType.IsValid() {
		return fmt.Errorf("invalid alias_type: %s", a.AliasType)
	}

	// Validate corporeal auto constraint
	if a.IsCorporealAuto && !a.IsAuto() {
		return fmt.Errorf("is_corporeal_auto can only be true for AUTO aliases")
	}

	return nil
}

// GetTTLDays returns TTL in days from metadata (for MASK aliases)
func (a *Alias) GetTTLDays() int {
	if a.Metadata == nil {
		return 0
	}

	if ttl, ok := a.Metadata["ttl_days"]; ok {
		switch v := ttl.(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
	}

	return 0
}

// SetTTLDays sets TTL in metadata
func (a *Alias) SetTTLDays(days int) {
	if a.Metadata == nil {
		a.Metadata = make(AliasMetadata)
	}
	a.Metadata["ttl_days"] = days
}

// GetDisplayName returns display name from metadata or alias_name
func (a *Alias) GetDisplayName() string {
	if a.Metadata == nil {
		return a.AliasName
	}

	if displayName, ok := a.Metadata["display_name"].(string); ok && displayName != "" {
		return displayName
	}

	return a.AliasName
}

// SetDisplayName sets display name in metadata
func (a *Alias) SetDisplayName(displayName string) {
	if a.Metadata == nil {
		a.Metadata = make(AliasMetadata)
	}
	a.Metadata["display_name"] = displayName
}

// AliasListResponse represents a list of aliases for API responses
type AliasListResponse struct {
	Aliases []Alias `json:"aliases"`
	Count   int     `json:"count"`
}

// GroupByType groups aliases by their type
func (alr *AliasListResponse) GroupByType() map[AliasType][]Alias {
	grouped := make(map[AliasType][]Alias)

	for _, alias := range alr.Aliases {
		grouped[alias.AliasType] = append(grouped[alias.AliasType], alias)
	}

	return grouped
}
