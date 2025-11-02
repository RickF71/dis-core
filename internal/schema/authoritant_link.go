package schema

import (
	"fmt"
	"time"
)

// AuthoritantLink — authoritant.link.v1
// Represents a verified, non-hierarchical relationship between multiple Authoritants.
// Used for federated consent, cross-domain validation, and recognition fields.
type AuthoritantLink struct {
	LinkID     string            `json:"link_id" yaml:"link_id"`
	Members    []string          `json:"members" yaml:"members"`
	Initiator  string            `json:"initiator" yaml:"initiator"`
	Signatures map[string]string `json:"signatures,omitempty" yaml:"signatures,omitempty"`
	LinkType   string            `json:"link_type" yaml:"link_type"` // peer | federation | relay
	TTL        string            `json:"ttl,omitempty" yaml:"ttl,omitempty"`
	ProofHash  string            `json:"proof_hash" yaml:"proof_hash"`
	CreatedAt  string            `json:"created_at" yaml:"created_at"`
	ExpiresAt  string            `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
	Status     string            `json:"status,omitempty" yaml:"status,omitempty"` // forming | active | expired | revoked
	Notes      string            `json:"notes,omitempty" yaml:"notes,omitempty"`
}

// Validate checks the structural integrity and logical consistency of an AuthoritantLink.
func (l AuthoritantLink) Validate() error {
	if err := requireNonEmpty("link_id", l.LinkID); err != nil {
		return err
	}
	if len(l.Members) < 2 {
		return fmt.Errorf("members: at least two Authoritants required to form a link")
	}
	if err := requireNonEmpty("initiator", l.Initiator); err != nil {
		return err
	}
	if err := requireNonEmpty("link_type", l.LinkType); err != nil {
		return err
	}
	if err := oneOf("link_type", l.LinkType, "peer", "federation", "relay"); err != nil {
		return err
	}
	if err := requireNonEmpty("proof_hash", l.ProofHash); err != nil {
		return err
	}
	if err := requireNonEmpty("created_at", l.CreatedAt); err != nil {
		return err
	}
	if err := ISO8601(l.CreatedAt); err != nil {
		return err
	}
	if l.ExpiresAt != "" {
		if err := ISO8601(l.ExpiresAt); err != nil {
			return err
		}
	}
	if l.Status != "" {
		if err := oneOf("status", l.Status, "forming", "active", "expired", "revoked"); err != nil {
			return err
		}
	}
	if l.TTL != "" {
		if _, err := time.ParseDuration(l.TTL); err != nil {
			// not fatal, but warn if invalid
			fmt.Printf("⚠️  invalid TTL format in AuthoritantLink %s: %v\n", l.LinkID, err)
		}
	}
	return nil
}
