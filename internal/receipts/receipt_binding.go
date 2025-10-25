package ledger

import (
	"time"
)

// SchemaBinding links a receipt to the specific schema and version
// it was validated or generated under.
type SchemaBinding struct {
	SchemaID string `json:"schema_id"`
	Version  string `json:"version"`
	Hash     string `json:"hash"`
}

// ReceiptBinding represents a lightweight verification record
// used internally for schema/domain validation.
// It does NOT contain cryptographic provenance — that’s handled by receipt.go.
type ReceiptBinding struct {
	ID        string        `json:"id"`
	When      time.Time     `json:"time"`
	Actor     string        `json:"actor"`
	Action    string        `json:"action"`
	PolicyRef string        `json:"policy_ref"`
	Binding   SchemaBinding `json:"binding"`
}

// Validate ensures the schema referenced by this binding exists
// and passes integrity verification (hash check) via the provided callback.
func (r *ReceiptBinding) Validate(verify func(id, version string) error) error {
	if verify == nil {
		return nil
	}
	return verify(r.Binding.SchemaID, r.Binding.Version)
}
