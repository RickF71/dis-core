package phase9c

import "time"

// Receipt represents a DIS receipt for Phase 9C provenance and verification
type Receipt struct {
	ID           string         `json:"id"`
	ReceiptType  string         `json:"receipt_type"`
	EventID      string         `json:"event_id"`
	PolicyRef    string         `json:"policy_ref"`
	RedactionRef string         `json:"redaction_ref"`
	IssuedBy     string         `json:"issued_by"`
	IssuedAt     time.Time      `json:"issued_at"`
	Verified     bool           `json:"verified"`
	Metadata     map[string]any `json:"metadata"`
}

// VerificationResult contains the result of receipt verification
type VerificationResult struct {
	ReceiptID    string   `json:"receipt_id"`
	Verified     bool     `json:"verified"`
	PolicyRef    string   `json:"policy_ref"`
	RedactionRef string   `json:"redaction_ref"`
	Timestamp    string   `json:"timestamp"`
	Issues       []string `json:"issues"`
}

// ReceiptStats contains statistics for the continuity dashboard
type ReceiptStats struct {
	Total     int            `json:"total"`
	Verified  int            `json:"verified"`
	Orphans   int            `json:"orphans"`
	ByType    map[string]int `json:"by_type"`
	Timestamp string         `json:"timestamp"`
}
