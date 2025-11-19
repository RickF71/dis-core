package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ChallengeStatus string

const (
	ChallengeStatusPending  ChallengeStatus = "pending"
	ChallengeStatusApproved ChallengeStatus = "approved"
	ChallengeStatusDenied   ChallengeStatus = "denied"
	ChallengeStatusExpired  ChallengeStatus = "expired"
)

// Challenge represents an external auth challenge created by FinaglerNone (or another external auth origin).
type Challenge struct {
	ID             uuid.UUID       `json:"id"`
	ExternalUserID uuid.UUID       `json:"external_user_id"`
	Status         ChallengeStatus `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
}

// ChallengeStore defines the storage interface for auth challenges.
type ChallengeStore interface {
	// GetChallenge retrieves a challenge by ID.
	GetChallenge(ctx context.Context, id uuid.UUID) (*Challenge, error)

	// SaveChallenge persists a new challenge.
	SaveChallenge(ctx context.Context, c *Challenge) error

	// Optional: extend later with status updates or deletes.
	// UpdateChallengeStatus(ctx context.Context, id uuid.UUID, status ChallengeStatus) error
	// DeleteChallenge(ctx context.Context, id uuid.UUID) error
}

// NewChallenge constructs a new pending challenge.
func NewChallenge(externalUserID uuid.UUID) *Challenge {
	return &Challenge{
		ID:             uuid.New(),
		ExternalUserID: externalUserID,
		Status:         ChallengeStatusPending,
		CreatedAt:      time.Now().UTC(),
	}
}
