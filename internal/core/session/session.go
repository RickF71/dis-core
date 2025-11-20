package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Session represents a persisted actor session
type Session struct {
	ID        string    `json:"id"`
	ActorID   string    `json:"actor_id"`
	DomainID  string    `json:"domain_id"`
	SeatID    string    `json:"seat_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewSessionToken returns a URL-safe base64 encoded 256-bit token
func NewSessionToken() (string, error) {
	b := make([]byte, 32) // 256-bit
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// CreateSessionTx inserts a session row inside the provided transaction.
// ttl is the desired duration until expiry (e.g., 8*time.Hour).
func CreateSessionTx(ctx context.Context, tx pgx.Tx, actorID, domainID, seatID string, ttl time.Duration) (string, string, error) {
	token, err := NewSessionToken()
	if err != nil {
		return "", "", err
	}
	id := uuid.New()
	now := time.Now().UTC()
	expires := now.Add(ttl)

	_, err = tx.Exec(ctx, `
        INSERT INTO sessions (id, actor_id, domain_id, seat_id, token, created_at, expires_at)
        VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7)
    `, id.String(), actorID, domainID, seatID, token, now, expires)
	if err != nil {
		return "", "", fmt.Errorf("create session: %w", err)
	}
	return id.String(), token, nil
}

// LoadSessionTx loads a session by token inside the provided transaction.
func LoadSessionTx(ctx context.Context, tx pgx.Tx, token string) (*Session, error) {
	var s Session
	row := tx.QueryRow(ctx, `
        SELECT id::text, actor_id::text, domain_id::text, seat_id::text, token, created_at, expires_at
        FROM sessions
        WHERE token = $1
        LIMIT 1
    `, token)
	if err := row.Scan(&s.ID, &s.ActorID, &s.DomainID, &s.SeatID, &s.Token, &s.CreatedAt, &s.ExpiresAt); err != nil {
		return nil, fmt.Errorf("load session: %w", err)
	}
	return &s, nil
}
