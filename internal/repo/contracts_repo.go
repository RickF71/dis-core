package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/contracts"
)

// GOV-13: Contracts Repository Implementation

// ContractRepository defines the interface for contract persistence operations
type ContractRepository interface {
	CreateContract(ctx context.Context, contract *contracts.Contract) (*contracts.Contract, error)
	GetContractByID(ctx context.Context, id string) (*contracts.Contract, error)
	ListContractsByDomain(ctx context.Context, domainID string, statusFilter string) ([]*contracts.Contract, error)
	ListContractsBySubject(ctx context.Context, subjectDomainID string, statusFilter string) ([]*contracts.Contract, error)
	UpdateContractStatus(ctx context.Context, id string, status contracts.ContractStatus, revokedAt *time.Time) (*contracts.Contract, error)
}

// contractRepository is the PostgreSQL implementation of ContractRepository
type contractRepository struct {
	pool *pgxpool.Pool
}

// NewContractRepository creates a new ContractRepository using a pgxpool.Pool
func NewContractRepository(pool *pgxpool.Pool) ContractRepository {
	return &contractRepository{pool: pool}
}

// CreateContract inserts a new contract into the database
func (r *contractRepository) CreateContract(ctx context.Context, contract *contracts.Contract) (*contracts.Contract, error) {
	if contract.ID == uuid.Nil {
		contract.ID = uuid.New()
	}
	if contract.CreatedAt.IsZero() {
		contract.CreatedAt = time.Now()
	}
	if contract.UpdatedAt.IsZero() {
		contract.UpdatedAt = time.Now()
	}

	// Compute status if not set
	if contract.Status == "" {
		contract.Status = contract.ComputeStatus()
	}

	// Validate before insert
	if err := contract.Validate(); err != nil {
		return nil, fmt.Errorf("contract validation failed: %w", err)
	}

	query := `
		INSERT INTO contracts (
			id, domain_id, subject_domain_id, alias_id,
			contract_type, dsci_channel, dsci_reference, dsci_version,
			effective_at, expires_at, revoked_at, status,
			created_at, updated_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		RETURNING id, domain_id, subject_domain_id, alias_id,
		          contract_type, dsci_channel, dsci_reference, dsci_version,
		          effective_at, expires_at, revoked_at, status,
		          created_at, updated_at, created_by
	`

	row := r.pool.QueryRow(ctx, query,
		contract.ID, contract.DomainID, contract.SubjectDomainID, contract.AliasID,
		contract.ContractType, contract.DSCIChannel, contract.DSCIReference, contract.DSCIVersion,
		contract.EffectiveAt, contract.ExpiresAt, contract.RevokedAt, contract.Status,
		contract.CreatedAt, contract.UpdatedAt, contract.CreatedBy,
	)

	var created contracts.Contract
	err := row.Scan(
		&created.ID, &created.DomainID, &created.SubjectDomainID, &created.AliasID,
		&created.ContractType, &created.DSCIChannel, &created.DSCIReference, &created.DSCIVersion,
		&created.EffectiveAt, &created.ExpiresAt, &created.RevokedAt, &created.Status,
		&created.CreatedAt, &created.UpdatedAt, &created.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert contract: %w", err)
	}

	return &created, nil
}

// GetContractByID retrieves a single contract by its ID
func (r *contractRepository) GetContractByID(ctx context.Context, id string) (*contracts.Contract, error) {
	contractID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid contract ID format: %w", err)
	}

	query := `
		SELECT id, domain_id, subject_domain_id, alias_id,
		       contract_type, dsci_channel, dsci_reference, dsci_version,
		       effective_at, expires_at, revoked_at, status,
		       created_at, updated_at, created_by
		FROM contracts
		WHERE id = $1
	`

	var contract contracts.Contract
	err = r.pool.QueryRow(ctx, query, contractID).Scan(
		&contract.ID, &contract.DomainID, &contract.SubjectDomainID, &contract.AliasID,
		&contract.ContractType, &contract.DSCIChannel, &contract.DSCIReference, &contract.DSCIVersion,
		&contract.EffectiveAt, &contract.ExpiresAt, &contract.RevokedAt, &contract.Status,
		&contract.CreatedAt, &contract.UpdatedAt, &contract.CreatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("contract not found: %s", id)
		}
		return nil, fmt.Errorf("failed to retrieve contract: %w", err)
	}

	return &contract, nil
}

// ListContractsByDomain retrieves all contracts issued by a specific domain
// Optional statusFilter: "active", "expired", "revoked", "pending", or "" for all
func (r *contractRepository) ListContractsByDomain(ctx context.Context, domainID string, statusFilter string) ([]*contracts.Contract, error) {
	domainUUID, err := uuid.Parse(domainID)
	if err != nil {
		return nil, fmt.Errorf("invalid domain ID format: %w", err)
	}

	query := `
		SELECT id, domain_id, subject_domain_id, alias_id,
		       contract_type, dsci_channel, dsci_reference, dsci_version,
		       effective_at, expires_at, revoked_at, status,
		       created_at, updated_at, created_by
		FROM contracts
		WHERE domain_id = $1
	`

	args := []interface{}{domainUUID}

	if statusFilter != "" {
		query += " AND status = $2"
		args = append(args, statusFilter)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list contracts by domain: %w", err)
	}
	defer rows.Close()

	var contractList []*contracts.Contract
	for rows.Next() {
		var contract contracts.Contract
		err := rows.Scan(
			&contract.ID, &contract.DomainID, &contract.SubjectDomainID, &contract.AliasID,
			&contract.ContractType, &contract.DSCIChannel, &contract.DSCIReference, &contract.DSCIVersion,
			&contract.EffectiveAt, &contract.ExpiresAt, &contract.RevokedAt, &contract.Status,
			&contract.CreatedAt, &contract.UpdatedAt, &contract.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contract row: %w", err)
		}
		contractList = append(contractList, &contract)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating contract rows: %w", rows.Err())
	}

	return contractList, nil
}

// ListContractsBySubject retrieves all contracts where a specific domain is the subject
// Optional statusFilter: "active", "expired", "revoked", "pending", or "" for all
func (r *contractRepository) ListContractsBySubject(ctx context.Context, subjectDomainID string, statusFilter string) ([]*contracts.Contract, error) {
	subjectUUID, err := uuid.Parse(subjectDomainID)
	if err != nil {
		return nil, fmt.Errorf("invalid subject domain ID format: %w", err)
	}

	query := `
		SELECT id, domain_id, subject_domain_id, alias_id,
		       contract_type, dsci_channel, dsci_reference, dsci_version,
		       effective_at, expires_at, revoked_at, status,
		       created_at, updated_at, created_by
		FROM contracts
		WHERE subject_domain_id = $1
	`

	args := []interface{}{subjectUUID}

	if statusFilter != "" {
		query += " AND status = $2"
		args = append(args, statusFilter)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list contracts by subject: %w", err)
	}
	defer rows.Close()

	var contractList []*contracts.Contract
	for rows.Next() {
		var contract contracts.Contract
		err := rows.Scan(
			&contract.ID, &contract.DomainID, &contract.SubjectDomainID, &contract.AliasID,
			&contract.ContractType, &contract.DSCIChannel, &contract.DSCIReference, &contract.DSCIVersion,
			&contract.EffectiveAt, &contract.ExpiresAt, &contract.RevokedAt, &contract.Status,
			&contract.CreatedAt, &contract.UpdatedAt, &contract.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contract row: %w", err)
		}
		contractList = append(contractList, &contract)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating contract rows: %w", rows.Err())
	}

	return contractList, nil
}

// UpdateContractStatus updates the status of a contract (used for revocation and expiry)
func (r *contractRepository) UpdateContractStatus(ctx context.Context, id string, status contracts.ContractStatus, revokedAt *time.Time) (*contracts.Contract, error) {
	contractID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid contract ID format: %w", err)
	}

	if !status.IsValid() {
		return nil, fmt.Errorf("invalid contract status: %s", status)
	}

	// Validate revoked_at consistency
	if status == contracts.ContractStatusRevoked && revokedAt == nil {
		return nil, fmt.Errorf("revoked_at is required when status is 'revoked'")
	}
	if status != contracts.ContractStatusRevoked && revokedAt != nil {
		return nil, fmt.Errorf("revoked_at can only be set when status is 'revoked'")
	}

	query := `
		UPDATE contracts
		SET status = $2, revoked_at = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, domain_id, subject_domain_id, alias_id,
		          contract_type, dsci_channel, dsci_reference, dsci_version,
		          effective_at, expires_at, revoked_at, status,
		          created_at, updated_at, created_by
	`

	var updated contracts.Contract
	err = r.pool.QueryRow(ctx, query, contractID, status, revokedAt).Scan(
		&updated.ID, &updated.DomainID, &updated.SubjectDomainID, &updated.AliasID,
		&updated.ContractType, &updated.DSCIChannel, &updated.DSCIReference, &updated.DSCIVersion,
		&updated.EffectiveAt, &updated.ExpiresAt, &updated.RevokedAt, &updated.Status,
		&updated.CreatedAt, &updated.UpdatedAt, &updated.CreatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("contract not found: %s", id)
		}
		return nil, fmt.Errorf("failed to update contract status: %w", err)
	}

	return &updated, nil
}
