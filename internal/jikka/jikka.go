package jikka

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Flow represents a single directional communication or intent between two participants.
type Flow struct {
	From      uuid.UUID      `json:"from"`
	To        uuid.UUID      `json:"to"`
	Intent    string         `json:"intent,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// Jikka represents the moral or relational bridge between two or more domains.
type Jikka struct {
	ID           uuid.UUID       `json:"id"`
	Participants []uuid.UUID     `json:"participants"`
	Flows        json.RawMessage `json:"flows"`
	Rules        json.RawMessage `json:"rules"`
	State        json.RawMessage `json:"state"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// GetJikkaByID retrieves a Jikka from the database by ID.
func GetJikkaByID(db *sql.DB, id uuid.UUID) (*Jikka, error) {
	var j Jikka
	row := db.QueryRow(`
		SELECT id, participants, flows, rules, state, created_at, updated_at
		FROM jikkas WHERE id = $1`, id)
	err := row.Scan(&j.ID, pq.Array(&j.Participants),
		&j.Flows, &j.Rules, &j.State, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// CreateJikka inserts a new Jikka into the database.
func CreateJikka(db *sql.DB, j *Jikka) (*Jikka, error) {
	// Ensure JSONB defaults — this fixes "invalid input syntax for type json"
	if j.Flows == nil {
		j.Flows = json.RawMessage(`{}`)
	}
	if j.Rules == nil {
		j.Rules = json.RawMessage(`{}`)
	}
	if j.State == nil {
		j.State = json.RawMessage(`{}`)
	}

	err := db.QueryRow(`
        INSERT INTO jikkas (participants, flows, rules, state)
		VALUES ($1::uuid[], $2, $3, $4)
		RETURNING id, created_at, updated_at`,
		pq.Array(j.Participants), j.Flows, j.Rules, j.State).
		Scan(&j.ID, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return j, nil
}
