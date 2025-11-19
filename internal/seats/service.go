package seats

import (
	"context"

	"github.com/google/uuid"
)

// Service provides business logic for seat operations
type Service struct {
	repo *Repository
}

// NewService creates a new seats service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// BuildSeatContext creates OPA input context from domain seats
func (s *Service) BuildSeatContext(ctx context.Context, domainID uuid.UUID) (*SeatContext, error) {
	seats, err := s.repo.GetActiveSeats(ctx, domainID)
	if err != nil {
		return &SeatContext{}, nil // Default to empty context on error
	}

	seatCtx := &SeatContext{}

	for _, seat := range seats {
		switch seat.SeatType {
		case "root":
			if seat.Status == "active" {
				seatCtx.ActiveRoot = true
			}
			if seat.Status == "frozen" {
				seatCtx.Frozen = true
			}
			if seat.Status == "detached" {
				seatCtx.Detached = true
			}

		case "member":
			if seat.Status == "active" {
				seatCtx.ActiveMember = true
			}
			if seat.Status == "frozen" {
				seatCtx.Frozen = true
			}
			if seat.Status == "detached" {
				seatCtx.Detached = true
			}
		}
	}

	return seatCtx, nil
}

// GetPerSeatRegoPolicies retrieves all active per-seat REGO policies for a domain
func (s *Service) GetPerSeatRegoPolicies(ctx context.Context, domainID uuid.UUID) (map[string]string, error) {
	seats, err := s.repo.GetActiveSeats(ctx, domainID)
	if err != nil {
		return nil, err
	}

	policies := make(map[string]string)
	for _, seat := range seats {
		if seat.RegoText != nil && *seat.RegoText != "" {
			// Key by seat ID for namespace isolation
			policies[seat.ID.String()] = *seat.RegoText
		}
	}

	return policies, nil
}
