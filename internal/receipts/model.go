package receipts

import "time"

// Receipt represents the legacy flat receipt shape used historically in the
// system. We maintain it for backward compatibility and provide helpers to
// wrap legacy receipts into the newer semantic ReceiptEnvelope form.
type Receipt struct {
	ID               string         `json:"id"`
	ReceiptType      string         `json:"receipt_type"`
	EventID          string         `json:"event_id"`
	PolicyRef        string         `json:"policy_ref"`
	RedactionRef     string         `json:"redaction_ref"`
	IssuedBy         string         `json:"issued_by"`
	IssuedAt         time.Time      `json:"issued_at"`
	Verified         bool           `json:"verified"`
	Metadata         map[string]any `json:"metadata"`
	OriginDomainID   string         `json:"origin_domain_id"`
	OriginDomainName string         `json:"origin_domain_name"`
}

// VerificationResult describes a simple verification outcome used by the
// Phase 9C receipt verification endpoint.
type VerificationResult struct {
	ReceiptID    string   `json:"receipt_id"`
	Verified     bool     `json:"verified"`
	PolicyRef    string   `json:"policy_ref"`
	RedactionRef string   `json:"redaction_ref"`
	Timestamp    string   `json:"timestamp"`
	Issues       []string `json:"issues"`
}

// WrapLegacyReceipt converts a legacy flat Receipt into a ReceiptEnvelope by
// preserving fields into the appropriate panels. We avoid importing domain
// types by accepting origin id and name as strings.
func WrapLegacyReceipt(originID string, originName string, r *Receipt) *ReceiptEnvelope {
	env := NewEnvelope(originID, originName, r.IssuedBy)

	// Move legacy metadata into the policy panel if present
	if r.Metadata != nil {
		for k, v := range r.Metadata {
			env.PolicyPanel[k] = v
		}
	}

	// Create a compact action payload with key legacy fields
	payload := map[string]any{
		"id":            r.ID,
		"receipt_type":  r.ReceiptType,
		"event_id":      r.EventID,
		"issued_by":     r.IssuedBy,
		"issued_at":     r.IssuedAt,
		"verified":      r.Verified,
		"redaction_ref": r.RedactionRef,
		"policy_ref":    r.PolicyRef,
	}
	env.ActionPanel = payload

	if originName != "" {
		env.DomainPanel["origin_name"] = originName
	}
	if originID != "" {
		env.DomainPanel["origin_id"] = originID
	}

	return env
}
