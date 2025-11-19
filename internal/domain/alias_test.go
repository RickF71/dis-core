package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// GOV-12: Alias Canon & DSCI Integration - Domain Model Tests

func TestAliasTypeIsValid(t *testing.T) {
	tests := []struct {
		name      string
		aliasType AliasType
		want      bool
	}{
		{"AUTO is valid", AliasTypeAuto, true},
		{"RELATIONSHIP is valid", AliasTypeRelationship, true},
		{"MASK is valid", AliasTypeMask, true},
		{"Empty string is invalid", AliasType(""), false},
		{"Random string is invalid", AliasType("RANDOM"), false},
		{"Lowercase is invalid", AliasType("auto"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.aliasType.IsValid(); got != tt.want {
				t.Errorf("AliasType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAliasStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		retiredAt  *time.Time
		wantStatus AliasStatus
	}{
		{"Active when retiredAt is nil", nil, AliasStatusActive},
		{"Retired when retiredAt is set", &now, AliasStatusRetired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias := &Alias{RetiredAt: tt.retiredAt}
			if got := alias.Status(); got != tt.wantStatus {
				t.Errorf("Alias.Status() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestAliasIsActive(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		retiredAt *time.Time
		want      bool
	}{
		{"Active when retiredAt is nil", nil, true},
		{"Not active when retiredAt is set", &now, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias := &Alias{RetiredAt: tt.retiredAt}
			if got := alias.IsActive(); got != tt.want {
				t.Errorf("Alias.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAliasTypeCheckers(t *testing.T) {
	autoAlias := &Alias{AliasType: AliasTypeAuto}
	relAlias := &Alias{AliasType: AliasTypeRelationship}
	maskAlias := &Alias{AliasType: AliasTypeMask}

	// Test IsAuto
	if !autoAlias.IsAuto() {
		t.Error("AUTO alias should return true for IsAuto()")
	}
	if relAlias.IsAuto() {
		t.Error("RELATIONSHIP alias should return false for IsAuto()")
	}
	if maskAlias.IsAuto() {
		t.Error("MASK alias should return false for IsAuto()")
	}

	// Test IsRelationship
	if autoAlias.IsRelationship() {
		t.Error("AUTO alias should return false for IsRelationship()")
	}
	if !relAlias.IsRelationship() {
		t.Error("RELATIONSHIP alias should return true for IsRelationship()")
	}
	if maskAlias.IsRelationship() {
		t.Error("MASK alias should return false for IsRelationship()")
	}

	// Test IsMask
	if autoAlias.IsMask() {
		t.Error("AUTO alias should return false for IsMask()")
	}
	if relAlias.IsMask() {
		t.Error("RELATIONSHIP alias should return false for IsMask()")
	}
	if !maskAlias.IsMask() {
		t.Error("MASK alias should return true for IsMask()")
	}
}

func TestAliasValidate(t *testing.T) {
	validID := uuid.New()
	ownerID := uuid.New()
	targetID := uuid.New()

	tests := []struct {
		name    string
		alias   *Alias
		wantErr bool
	}{
		{
			name: "Valid AUTO alias",
			alias: &Alias{
				ID:             validID,
				OwnerDomainID:  ownerID,
				TargetDomainID: targetID,
				AliasName:      "test-alias",
				AliasType:      AliasTypeAuto,
			},
			wantErr: false,
		},
		{
			name: "Valid RELATIONSHIP alias",
			alias: &Alias{
				ID:             validID,
				OwnerDomainID:  ownerID,
				TargetDomainID: targetID,
				AliasName:      "myhandle",
				AliasType:      AliasTypeRelationship,
			},
			wantErr: false,
		},
		{
			name: "Valid MASK alias",
			alias: &Alias{
				ID:             validID,
				OwnerDomainID:  ownerID,
				TargetDomainID: ownerID, // Self-targeting
				AliasName:      "mask-abc123",
				AliasType:      AliasTypeMask,
			},
			wantErr: false,
		},
		{
			name: "Missing ID",
			alias: &Alias{
				OwnerDomainID:  ownerID,
				TargetDomainID: targetID,
				AliasName:      "test",
				AliasType:      AliasTypeAuto,
			},
			wantErr: true,
		},
		{
			name: "Missing owner domain ID",
			alias: &Alias{
				ID:             validID,
				TargetDomainID: targetID,
				AliasName:      "test",
				AliasType:      AliasTypeAuto,
			},
			wantErr: true,
		},
		{
			name: "Missing target domain ID",
			alias: &Alias{
				ID:            validID,
				OwnerDomainID: ownerID,
				AliasName:     "test",
				AliasType:     AliasTypeAuto,
			},
			wantErr: true,
		},
		{
			name: "Missing alias name",
			alias: &Alias{
				ID:             validID,
				OwnerDomainID:  ownerID,
				TargetDomainID: targetID,
				AliasType:      AliasTypeAuto,
			},
			wantErr: true,
		},
		{
			name: "Invalid alias type",
			alias: &Alias{
				ID:             validID,
				OwnerDomainID:  ownerID,
				TargetDomainID: targetID,
				AliasName:      "test",
				AliasType:      AliasType("INVALID"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.alias.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Alias.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAliasGetSetTTLDays(t *testing.T) {
	alias := &Alias{
		AliasType: AliasTypeMask,
		Metadata:  make(AliasMetadata),
	}

	// Initially no TTL
	if ttl := alias.GetTTLDays(); ttl != 0 {
		t.Errorf("Expected TTL = 0, got %d", ttl)
	}

	// Set TTL
	alias.SetTTLDays(30)
	if ttl := alias.GetTTLDays(); ttl != 30 {
		t.Errorf("Expected TTL = 30, got %d", ttl)
	}

	// Set different TTL
	alias.SetTTLDays(90)
	if ttl := alias.GetTTLDays(); ttl != 90 {
		t.Errorf("Expected TTL = 90, got %d", ttl)
	}

	// Test with nil metadata
	aliasNoMetadata := &Alias{
		AliasType: AliasTypeMask,
		Metadata:  nil,
	}
	aliasNoMetadata.SetTTLDays(7)
	if ttl := aliasNoMetadata.GetTTLDays(); ttl != 7 {
		t.Errorf("Expected TTL = 7 with nil metadata, got %d", ttl)
	}
}

func TestAliasGetSetDisplayName(t *testing.T) {
	alias := &Alias{
		AliasName: "mask-abc123",
		Metadata:  make(AliasMetadata),
	}

	// Initially returns alias_name as fallback
	if name := alias.GetDisplayName(); name != "mask-abc123" {
		t.Errorf("Expected display name = 'mask-abc123', got '%s'", name)
	}

	// Set display name
	alias.SetDisplayName("My Browsing Persona")
	if name := alias.GetDisplayName(); name != "My Browsing Persona" {
		t.Errorf("Expected display name = 'My Browsing Persona', got '%s'", name)
	}

	// Change display name
	alias.SetDisplayName("Anonymous Explorer")
	if name := alias.GetDisplayName(); name != "Anonymous Explorer" {
		t.Errorf("Expected display name = 'Anonymous Explorer', got '%s'", name)
	}

	// Test with nil metadata
	aliasNoMetadata := &Alias{
		AliasName: "test-alias",
		Metadata:  nil,
	}
	if name := aliasNoMetadata.GetDisplayName(); name != "test-alias" {
		t.Errorf("Expected fallback to alias_name, got '%s'", name)
	}
	aliasNoMetadata.SetDisplayName("Custom Name")
	if name := aliasNoMetadata.GetDisplayName(); name != "Custom Name" {
		t.Errorf("Expected display name = 'Custom Name', got '%s'", name)
	}
}

func TestAliasListResponseGroupByType(t *testing.T) {
	aliases := []Alias{
		{AliasType: AliasTypeAuto, AliasName: "auto1"},
		{AliasType: AliasTypeAuto, AliasName: "auto2"},
		{AliasType: AliasTypeRelationship, AliasName: "rel1"},
		{AliasType: AliasTypeMask, AliasName: "mask1"},
		{AliasType: AliasTypeMask, AliasName: "mask2"},
		{AliasType: AliasTypeMask, AliasName: "mask3"},
	}

	resp := AliasListResponse{Aliases: aliases}
	grouped := resp.GroupByType()

	// Check AUTO aliases
	if len(grouped[AliasTypeAuto]) != 2 {
		t.Errorf("Expected 2 AUTO aliases, got %d", len(grouped[AliasTypeAuto]))
	}

	// Check RELATIONSHIP aliases
	if len(grouped[AliasTypeRelationship]) != 1 {
		t.Errorf("Expected 1 RELATIONSHIP alias, got %d", len(grouped[AliasTypeRelationship]))
	}

	// Check MASK aliases
	if len(grouped[AliasTypeMask]) != 3 {
		t.Errorf("Expected 3 MASK aliases, got %d", len(grouped[AliasTypeMask]))
	}
}

func TestAliasMetadataJSONMarshaling(t *testing.T) {
	metadata := AliasMetadata{
		"ttl_days":     30,
		"display_name": "Test Alias",
		"custom_field": "custom value",
	}

	alias := &Alias{
		ID:             uuid.New(),
		OwnerDomainID:  uuid.New(),
		TargetDomainID: uuid.New(),
		AliasName:      "test",
		AliasType:      AliasTypeMask,
		Metadata:       metadata,
	}

	// Verify metadata fields are accessible
	if alias.GetTTLDays() != 30 {
		t.Errorf("Expected TTL = 30, got %d", alias.GetTTLDays())
	}
	if alias.GetDisplayName() != "Test Alias" {
		t.Errorf("Expected display_name = 'Test Alias', got '%s'", alias.GetDisplayName())
	}
	if alias.Metadata["custom_field"] != "custom value" {
		t.Errorf("Expected custom_field = 'custom value', got '%v'", alias.Metadata["custom_field"])
	}
}
