package model

import "time"

type DisEventType string

type DisEvent struct {
	ID        int          `db:"id" json:"id"`
	TS        time.Time    `db:"ts" json:"ts"`
	Type      DisEventType `db:"type" json:"type"`
	Actor     string       `db:"actor" json:"actor"`
	Payload   []byte       `db:"payload" json:"payload"`
	Signature string       `db:"sig" json:"sig"`
}
