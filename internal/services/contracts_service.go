package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dis-core/internal/contracts"
	"dis-core/internal/receipts"
	"dis-core/internal/repo"
)

// GOV-13: Contract Service Implementation

// ContractService defines business logic for contract operations
type ContractService interface {
	CreateContract(ctx context.Context, input contracts.CreateContractInput) (*contracts.Contract, error)
	RevokeContract(ctx context.Context, contractID string, reason string, actorID string) (*contracts.Contract, error)
	GetContract(ctx context.Context, contractID string) (*contracts.Contract, error)
	ListDomainContracts(ctx context.Context, domainID string, statusFilter string) ([]*contracts.Contract, error)
	ListSubjectContracts(ctx context.Context, subjectDomainID string, statusFilter string) ([]*contracts.Contract, error)
}

// contractService implements ContractService
type contractService struct {
	contractRepo    repo.ContractRepository
	receiptRecorder receipts.ReceiptRecorder
}

// NewContractService creates a new ContractService
func NewContractService(contractRepo repo.ContractRepository, receiptRecorder receipts.ReceiptRecorder) ContractService {
	return &contractService{
		contractRepo:    contractRepo,
		receiptRecorder: receiptRecorder,
	}
}

// CreateContract creates a new contract and records a receipt
func (s *contractService) CreateContract(ctx context.Context, input contracts.CreateContractInput) (*contracts.Contract, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid contract input: %w", err)
	}

	// Convert input to contract with computed status
	contract := input.ToContract()

	// Persist contract
	created, err := s.contractRepo.CreateContract(ctx, contract)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	// Record receipt (non-blocking, best-effort)
	if s.receiptRecorder != nil {
		go func() {
			receiptCtx := context.Background() // Detached context
			if err := receipts.RecordContractCreateReceipt(receiptCtx, s.receiptRecorder, created); err != nil {
				// Log but don't fail the contract creation
				fmt.Printf("Warning: failed to record contract create receipt: %v\n", err)
			}
		}()
	}

	return created, nil
}

// RevokeContract marks a contract as revoked and records a receipt
func (s *contractService) RevokeContract(ctx context.Context, contractID string, reason string, actorID string) (*contracts.Contract, error) {
	// Retrieve existing contract
	existing, err := s.contractRepo.GetContractByID(ctx, contractID)
	if err != nil {
		return nil, fmt.Errorf("contract not found: %w", err)
	}

	// Idempotency: If already revoked, return existing
	if existing.IsRevoked() {
		return existing, nil
	}

	// Update contract status to revoked
	now := time.Now()
	revoked, err := s.contractRepo.UpdateContractStatus(ctx, contractID, contracts.ContractStatusRevoked, &now)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke contract: %w", err)
	}

	// Record revocation receipt (non-blocking, best-effort)
	if s.receiptRecorder != nil {
		go func() {
			receiptCtx := context.Background() // Detached context
			if err := receipts.RecordContractRevokeReceipt(receiptCtx, s.receiptRecorder, revoked, reason, actorID); err != nil {
				fmt.Printf("Warning: failed to record contract revoke receipt: %v\n", err)
			}
		}()
	}

	return revoked, nil
}

// GetContract retrieves a single contract by ID
func (s *contractService) GetContract(ctx context.Context, contractID string) (*contracts.Contract, error) {
	contract, err := s.contractRepo.GetContractByID(ctx, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve contract: %w", err)
	}
	return contract, nil
}

// ListDomainContracts lists all contracts issued by a specific domain
func (s *contractService) ListDomainContracts(ctx context.Context, domainID string, statusFilter string) ([]*contracts.Contract, error) {
	// Validate domain ID format
	if _, err := uuid.Parse(domainID); err != nil {
		return nil, fmt.Errorf("invalid domain ID format: %w", err)
	}

	// Validate status filter if provided
	if statusFilter != "" {
		validStatus := contracts.ContractStatus(statusFilter)
		if !validStatus.IsValid() {
			return nil, fmt.Errorf("invalid status filter: %s", statusFilter)
		}
	}

	contractList, err := s.contractRepo.ListContractsByDomain(ctx, domainID, statusFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list domain contracts: %w", err)
	}

	return contractList, nil
}

// ListSubjectContracts lists all contracts where a specific domain is the subject
func (s *contractService) ListSubjectContracts(ctx context.Context, subjectDomainID string, statusFilter string) ([]*contracts.Contract, error) {
	// Validate subject domain ID format
	if _, err := uuid.Parse(subjectDomainID); err != nil {
		return nil, fmt.Errorf("invalid subject domain ID format: %w", err)
	}

	// Validate status filter if provided
	if statusFilter != "" {
		validStatus := contracts.ContractStatus(statusFilter)
		if !validStatus.IsValid() {
			return nil, fmt.Errorf("invalid status filter: %s", statusFilter)
		}
	}

	contractList, err := s.contractRepo.ListContractsBySubject(ctx, subjectDomainID, statusFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list subject contracts: %w", err)
	}

	return contractList, nil
}
