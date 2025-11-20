package services

import (
	"context"
	"fmt"

	"dis-core/internal/domain"
	"dis-core/internal/repo"

	"github.com/google/uuid"
)

// GOV-12: Alias Service
// Semantic enforcement for AUTO / RELATIONSHIP / MASK alias creation

// AliasService handles alias business logic
type AliasService struct {
	repo *repo.AliasRepository
}

// NewAliasService creates a new alias service
func NewAliasService(repo *repo.AliasRepository) *AliasService {
	return &AliasService{repo: repo}
}

// EnsureCorporealRootAlias ensures an AUTO alias exists for a corporeal domain
// This is called during corporeal domain creation/first rooting
func (s *AliasService) EnsureCorporealRootAlias(ctx context.Context, corporealDomainID uuid.UUID) (*domain.Alias, error) {
	// Check if alias already exists
	existing, err := s.repo.GetCorporealAutoAlias(ctx, corporealDomainID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing corporeal auto alias: %w", err)
	}

	if existing != nil {
		return existing, nil // Already exists
	}

	// Create new corporeal AUTO alias
	alias := &domain.Alias{
		OwnerDomainID:   corporealDomainID,
		TargetDomainID:  corporealDomainID, // Self-referential for corporeal root
		AliasName:       fmt.Sprintf("corporeal-auto-%s", corporealDomainID.String()[:8]),
		AliasType:       domain.AliasTypeAuto,
		IsCorporealAuto: true,
		Metadata:        make(domain.AliasMetadata),
	}

	if err := s.repo.CreateAlias(ctx, alias); err != nil {
		return nil, fmt.Errorf("failed to create corporeal auto alias: %w", err)
	}

	return alias, nil
}

// EnsureStructuralAlias creates an AUTO alias for structural membership
// This represents non-volitional graph binding (e.g., domain membership via seat)
func (s *AliasService) EnsureStructuralAlias(ctx context.Context, ownerDomainID, targetDomainID uuid.UUID) (*domain.Alias, error) {
	// Check if alias already exists
	aliasName := fmt.Sprintf("auto-%s", targetDomainID.String()[:8])
	existing, err := s.repo.FindActiveAliasByName(ctx, ownerDomainID, targetDomainID, aliasName)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing structural alias: %w", err)
	}

	if existing != nil {
		return existing, nil
	}

	// Create new structural AUTO alias
	alias := &domain.Alias{
		OwnerDomainID:   ownerDomainID,
		TargetDomainID:  targetDomainID,
		AliasName:       aliasName,
		AliasType:       domain.AliasTypeAuto,
		IsCorporealAuto: false,
		Metadata:        make(domain.AliasMetadata),
	}

	if err := s.repo.CreateAlias(ctx, alias); err != nil {
		return nil, fmt.Errorf("failed to create structural alias: %w", err)
	}

	return alias, nil
}

// CreateRelationshipAlias creates a volitional RELATIONSHIP alias
// This represents user-chosen identity within a specific domain
func (s *AliasService) CreateRelationshipAlias(
	ctx context.Context,
	ownerDomainID, targetDomainID uuid.UUID,
	aliasName string,
	dsciContractID *uuid.UUID,
	metadata domain.AliasMetadata,
) (*domain.Alias, error) {
	if aliasName == "" {
		return nil, fmt.Errorf("alias_name is required for RELATIONSHIP aliases")
	}

	// Check for conflicts
	existing, err := s.repo.FindActiveAliasByName(ctx, ownerDomainID, targetDomainID, aliasName)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing alias: %w", err)
	}

	if existing != nil {
		// Return the canonical conflict error string expected by API handlers/tests
		return nil, fmt.Errorf("alias already exists for this owner, target, and name")
	}

	// Create RELATIONSHIP alias
	if metadata == nil {
		metadata = make(domain.AliasMetadata)
	}

	alias := &domain.Alias{
		OwnerDomainID:  ownerDomainID,
		TargetDomainID: targetDomainID,
		AliasName:      aliasName,
		AliasType:      domain.AliasTypeRelationship,
		DSCIContractID: dsciContractID,
		Metadata:       metadata,
	}

	if err := s.repo.CreateAlias(ctx, alias); err != nil {
		return nil, fmt.Errorf("failed to create relationship alias: %w", err)
	}

	return alias, nil
}

// CreateMaskAlias creates an ephemeral MASK alias for incognito browsing
// These have minimal authority by default and can be short-lived
func (s *AliasService) CreateMaskAlias(
	ctx context.Context,
	ownerDomainID uuid.UUID,
	requestedName string,
	ttlDays int,
	metadata domain.AliasMetadata,
) (*domain.Alias, error) {
	// Generate mask name if not provided
	aliasName := requestedName
	if aliasName == "" {
		aliasName = repo.GenerateMaskName()
	}

	// Initialize metadata
	if metadata == nil {
		metadata = make(domain.AliasMetadata)
	}

	// Set TTL if specified
	if ttlDays > 0 {
		metadata["ttl_days"] = ttlDays
	}

	// MASK aliases are self-targeting
	alias := &domain.Alias{
		OwnerDomainID:  ownerDomainID,
		TargetDomainID: ownerDomainID,
		AliasName:      aliasName,
		AliasType:      domain.AliasTypeMask,
		Metadata:       metadata,
	}

	if err := s.repo.CreateAlias(ctx, alias); err != nil {
		return nil, fmt.Errorf("failed to create mask alias: %w", err)
	}

	return alias, nil
}

// RetireAlias retires an alias with a reason
func (s *AliasService) RetireAlias(ctx context.Context, aliasID uuid.UUID, reason string) error {
	return s.repo.RetireAlias(ctx, aliasID, reason)
}

// GetAliasesForDomain retrieves all aliases for a domain (owner or target)
func (s *AliasService) GetAliasesForDomain(ctx context.Context, domainID uuid.UUID, includeRetired bool) ([]domain.Alias, error) {
	// Get aliases where domain is owner
	ownerAliases, err := s.repo.GetAliasesForOwner(ctx, domainID, includeRetired)
	if err != nil {
		return nil, err
	}

	// Get aliases where domain is target
	targetAliases, err := s.repo.GetAliasesForTarget(ctx, domainID, includeRetired)
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate
	aliasMap := make(map[uuid.UUID]domain.Alias)
	for _, a := range ownerAliases {
		aliasMap[a.ID] = a
	}
	for _, a := range targetAliases {
		aliasMap[a.ID] = a
	}

	aliases := make([]domain.Alias, 0, len(aliasMap))
	for _, a := range aliasMap {
		aliases = append(aliases, a)
	}

	return aliases, nil
}
