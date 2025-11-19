package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresChallengeStore struct {
	DB *pgxpool.Pool
}

// NewPostgresChallengeStore constructs a new challenge store.
func NewPostgresChallengeStore(db *pgxpool.Pool) *PostgresChallengeStore {
	return &PostgresChallengeStore{DB: db}
}

func (s *PostgresChallengeStore) SaveChallenge(ctx context.Context, c *Challenge) error {
	const q = `
        INSERT INTO auth_challenges (id, external_user_id, status, created_at)
        VALUES ($1, $2, $3, $4)
    `
	_, err := s.DB.Exec(ctx, q, c.ID, c.ExternalUserID, string(c.Status), c.CreatedAt)
	return err
}

func (s *PostgresChallengeStore) GetChallenge(ctx context.Context, id uuid.UUID) (*Challenge, error) {
	const q = `
        SELECT id, external_user_id, status, created_at
        FROM auth_challenges
        WHERE id = $1
    `
	var (
		challengeID    uuid.UUID
		externalUserID uuid.UUID
		status         string
		createdAt      time.Time
	)

	err := s.DB.QueryRow(ctx, q, id).Scan(&challengeID, &externalUserID, &status, &createdAt)
	if err != nil {
		return nil, err
	}

	c := &Challenge{
		ID:             challengeID,
		ExternalUserID: externalUserID,
		Status:         ChallengeStatus(status),
		CreatedAt:      createdAt,
	}
	return c, nil
}

// Optional: implement UpdateChallengeStatus later if/when needed.
// func (s *PostgresChallengeStore) UpdateChallengeStatus(ctx context.Context, id uuid.UUID, status ChallengeStatus) error {
//     const q = `UPDATE auth_challenges SET status = $1 WHERE id = $2`
//     _, err := s.DB.Exec(ctx, q, string(status), id)
//     return err
// }
