package schema

import "fmt"

// SharedCoreTriad — jikka.shared-core.triad.v1
// Parent + Authorization (DomAuth) + Child, bound by a Shared Core domain.
type SharedCoreTriad struct {
	Type             string `json:"type" yaml:"type"`                         // jikka.shared-core.triad.v1
	Version          string `json:"version" yaml:"version"`
	Domain           string `json:"domain" yaml:"domain"`                     // shared core domain id
	ParentDomain     string `json:"parent_domain" yaml:"parent_domain"`
	AuthorizationDom string `json:"authorization_domain" yaml:"authorization_domain"`
	ChildDomain      string `json:"child_domain" yaml:"child_domain"`
	SharedCoreDomain string `json:"shared_core_domain" yaml:"shared_core_domain"`
	Threshold        int    `json:"threshold" yaml:"threshold"`
	ConsentProof     string `json:"consent_proof,omitempty" yaml:"consent_proof,omitempty"`
	CreatedAt        string `json:"created_at" yaml:"created_at"`
	Status           string `json:"status,omitempty" yaml:"status,omitempty"` // forming | active | severed
}

func (s SharedCoreTriad) Validate() error {
	if err := requireNonEmpty("type", s.Type); err != nil { return err }
	if err := requireNonEmpty("version", s.Version); err != nil { return err }
	if err := requireNonEmpty("domain", s.Domain); err != nil { return err }
	if err := requireNonEmpty("parent_domain", s.ParentDomain); err != nil { return err }
	if err := requireNonEmpty("authorization_domain", s.AuthorizationDom); err != nil { return err }
	if err := requireNonEmpty("child_domain", s.ChildDomain); err != nil { return err }
	if err := requireNonEmpty("shared_core_domain", s.SharedCoreDomain); err != nil { return err }
	if s.Threshold < 3 { return fmt.Errorf("threshold must be >= 3 for a stable triad") }
	if err := requireNonEmpty("created_at", s.CreatedAt); err != nil { return err }
	if err := ISO8601(s.CreatedAt); err != nil { return err }
	if s.Status != "" {
		if err := oneOf("status", s.Status, "forming", "active", "severed"); err != nil { return err }
	}
	return nil
}
