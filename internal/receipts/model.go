package receipts

import "time"

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

// VerificationResult contains the result of Phase 9C receipt verification
type VerificationResult struct {
	ReceiptID    string   `json:"receipt_id"`
	Verified     bool     `json:"verified"`
	PolicyRef    string   `json:"policy_ref"`
	RedactionRef string   `json:"redaction_ref"`
	Timestamp    string   `json:"timestamp"`
	Issues       []string `json:"issues"`
}
