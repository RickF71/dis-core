package events

import "time"

// DisEventType is a string alias so you can type-flip later if needed.
type DisEventType string

// DisEvent is the append-only truth unit (the permanent, canonical record).
type DisEvent struct {
	ID        int          `db:"id" json:"id"`
	TS        time.Time    `db:"ts" json:"ts"`
	Type      DisEventType `db:"type" json:"type"`
	Actor     string       `db:"actor" json:"actor"`
	Payload   []byte       `db:"payload" json:"payload"` // raw JSON blob
	Signature string       `db:"sig" json:"sig"`         // placeholder for now
}
