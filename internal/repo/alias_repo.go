package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"dis-core/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-12: Alias Repository
// CRUD operations for canonical alias model

// AliasRepository handles alias persistence
type AliasRepository struct {
	pool *pgxpool.Pool
}

// NewAliasRepository creates a new alias repository
func NewAliasRepository(pool *pgxpool.Pool) *AliasRepository {
	return &AliasRepository{pool: pool}
}

// CreateAlias inserts a new alias
func (r *AliasRepository) CreateAlias(ctx context.Context, alias *domain.Alias) error {
	if err := alias.Validate(); err != nil {
		return fmt.Errorf("invalid alias: %w", err)
	}

	// Ensure ID is set
	if alias.ID == uuid.Nil {
		alias.ID = uuid.New()
	}

	// Ensure created_at is set
	if alias.CreatedAt.IsZero() {
		alias.CreatedAt = time.Now().UTC()
	}

	// Initialize metadata if nil
	if alias.Metadata == nil {
		alias.Metadata = make(domain.AliasMetadata)
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO aliases (
			id, owner_domain_id, target_domain_id, alias_name, alias_type,
			is_corporeal_auto, dsci_contract_id, created_at, retired_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, alias.ID, alias.OwnerDomainID, alias.TargetDomainID, alias.AliasName,
		alias.AliasType, alias.IsCorporealAuto, alias.DSCIContractID,
		alias.CreatedAt, alias.RetiredAt, alias.Metadata)

	if err != nil {
		return fmt.Errorf("failed to create alias: %w", err)
	}

	return nil
}

// RetireAlias marks an alias as retired
func (r *AliasRepository) RetireAlias(ctx context.Context, id uuid.UUID, reason string) error {
	now := time.Now().UTC()

	result, err := r.pool.Exec(ctx, `
		UPDATE aliases
		SET retired_at = $1,
		    metadata = jsonb_set(
				COALESCE(metadata, '{}'::jsonb),
				'{retirement_reason}',
				to_jsonb($2::text)
			)
		WHERE id = $3 AND retired_at IS NULL
	`, now, reason, id)

	if err != nil {
		return fmt.Errorf("failed to retire alias: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("alias not found or already retired: %s", id)
	}

	return nil
}

// GetAliasesForOwner retrieves all aliases owned by a domain
func (r *AliasRepository) GetAliasesForOwner(ctx context.Context, ownerDomainID uuid.UUID, includeRetired bool) ([]domain.Alias, error) {
	query := `
		SELECT id, owner_domain_id, target_domain_id, alias_name, alias_type,
		       is_corporeal_auto, dsci_contract_id, created_at, retired_at, metadata
		FROM aliases
		WHERE owner_domain_id = $1
	`

	if !includeRetired {
		query += " AND retired_at IS NULL"
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, ownerDomainID)
	if err != nil {
		return nil, fmt.Errorf("failed to query aliases for owner: %w", err)
	}
	defer rows.Close()

	return r.scanAliases(rows)
}

// GetAliasesForTarget retrieves all aliases targeting a domain
func (r *AliasRepository) GetAliasesForTarget(ctx context.Context, targetDomainID uuid.UUID, includeRetired bool) ([]domain.Alias, error) {
	query := `
		SELECT id, owner_domain_id, target_domain_id, alias_name, alias_type,
		       is_corporeal_auto, dsci_contract_id, created_at, retired_at, metadata
		FROM aliases
		WHERE target_domain_id = $1
	`

	if !includeRetired {
		query += " AND retired_at IS NULL"
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, targetDomainID)
	if err != nil {
		return nil, fmt.Errorf("failed to query aliases for target: %w", err)
	}
	defer rows.Close()

	return r.scanAliases(rows)
}

// GetAliasesByType retrieves aliases by type for an owner
func (r *AliasRepository) GetAliasesByType(ctx context.Context, ownerDomainID uuid.UUID, aliasType domain.AliasType, includeRetired bool) ([]domain.Alias, error) {
	query := `
		SELECT id, owner_domain_id, target_domain_id, alias_name, alias_type,
		       is_corporeal_auto, dsci_contract_id, created_at, retired_at, metadata
		FROM aliases
		WHERE owner_domain_id = $1 AND alias_type = $2
	`

	if !includeRetired {
		query += " AND retired_at IS NULL"
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, ownerDomainID, aliasType)
	if err != nil {
		return nil, fmt.Errorf("failed to query aliases by type: %w", err)
	}
	defer rows.Close()

	return r.scanAliases(rows)
}

// GetActiveMaskAliasesForOwner retrieves active MASK aliases for an owner
func (r *AliasRepository) GetActiveMaskAliasesForOwner(ctx context.Context, ownerDomainID uuid.UUID) ([]domain.Alias, error) {
	return r.GetAliasesByType(ctx, ownerDomainID, domain.AliasTypeMask, false)
}

// GetCorporealAutoAlias retrieves the corporeal AUTO alias for a domain
func (r *AliasRepository) GetCorporealAutoAlias(ctx context.Context, ownerDomainID uuid.UUID) (*domain.Alias, error) {
	var alias domain.Alias

	err := r.pool.QueryRow(ctx, `
		SELECT id, owner_domain_id, target_domain_id, alias_name, alias_type,
		       is_corporeal_auto, dsci_contract_id, created_at, retired_at, metadata
		FROM aliases
		WHERE owner_domain_id = $1
		  AND alias_type = $2
		  AND is_corporeal_auto = true
		  AND retired_at IS NULL
		LIMIT 1
	`, ownerDomainID, domain.AliasTypeAuto).Scan(
		&alias.ID, &alias.OwnerDomainID, &alias.TargetDomainID, &alias.AliasName,
		&alias.AliasType, &alias.IsCorporealAuto, &alias.DSCIContractID,
		&alias.CreatedAt, &alias.RetiredAt, &alias.Metadata,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // Not found, but not an error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get corporeal auto alias: %w", err)
	}

	return &alias, nil
}

// GetAliasByID retrieves an alias by its ID
func (r *AliasRepository) GetAliasByID(ctx context.Context, id uuid.UUID) (*domain.Alias, error) {
	var alias domain.Alias

	err := r.pool.QueryRow(ctx, `
		SELECT id, owner_domain_id, target_domain_id, alias_name, alias_type,
		       is_corporeal_auto, dsci_contract_id, created_at, retired_at, metadata
		FROM aliases
		WHERE id = $1
	`, id).Scan(
		&alias.ID, &alias.OwnerDomainID, &alias.TargetDomainID, &alias.AliasName,
		&alias.AliasType, &alias.IsCorporealAuto, &alias.DSCIContractID,
		&alias.CreatedAt, &alias.RetiredAt, &alias.Metadata,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("alias not found: %s", id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get alias: %w", err)
	}

	return &alias, nil
}

// FindActiveAliasByName finds an active alias by name and owner/target
func (r *AliasRepository) FindActiveAliasByName(ctx context.Context, ownerDomainID, targetDomainID uuid.UUID, aliasName string) (*domain.Alias, error) {
	var alias domain.Alias

	err := r.pool.QueryRow(ctx, `
		SELECT id, owner_domain_id, target_domain_id, alias_name, alias_type,
		       is_corporeal_auto, dsci_contract_id, created_at, retired_at, metadata
		FROM aliases
		WHERE owner_domain_id = $1
		  AND target_domain_id = $2
		  AND alias_name = $3
		  AND retired_at IS NULL
		LIMIT 1
	`, ownerDomainID, targetDomainID, aliasName).Scan(
		&alias.ID, &alias.OwnerDomainID, &alias.TargetDomainID, &alias.AliasName,
		&alias.AliasType, &alias.IsCorporealAuto, &alias.DSCIContractID,
		&alias.CreatedAt, &alias.RetiredAt, &alias.Metadata,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // Not found
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find alias: %w", err)
	}

	return &alias, nil
}

// scanAliases scans multiple alias rows
func (r *AliasRepository) scanAliases(rows pgx.Rows) ([]domain.Alias, error) {
	var aliases []domain.Alias

	for rows.Next() {
		var alias domain.Alias
		err := rows.Scan(
			&alias.ID, &alias.OwnerDomainID, &alias.TargetDomainID, &alias.AliasName,
			&alias.AliasType, &alias.IsCorporealAuto, &alias.DSCIContractID,
			&alias.CreatedAt, &alias.RetiredAt, &alias.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alias: %w", err)
		}
		aliases = append(aliases, alias)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating aliases: %w", err)
	}

	return aliases, nil
}

// generateMaskName generates a random mask alias name
func generateMaskName() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return "mask-" + hex.EncodeToString(bytes)
}

// GenerateMaskName is exported for use in service layer
func GenerateMaskName() string {
	return generateMaskName()
}
