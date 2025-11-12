package policy

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
)

// Domain and Identity are assumed to exist in your models package.
type Domain struct {
	ID       string                 `json:"id"`
	ParentID string                 `json:"parent_id"`
	Name     string                 `json:"name"`
	State    string                 `json:"state"`
	Payload  map[string]interface{} `json:"payload,omitempty"`
}

type Identity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PolicyInput is the canonical self-aware envelope for every OPA call.
type PolicyInput struct {
	Self   Domain      `json:"self"`
	Parent *Domain     `json:"parent,omitempty"`
	Action interface{} `json:"action,omitempty"`
	Caller *Identity   `json:"caller,omitempty"`
	Chain  []string    `json:"chain"`
}

// BuildPolicyInput assembles a complete context for a given domain.
func BuildPolicyInput(domain *Domain) (*PolicyInput, error) {
	if domain == nil {
		return nil, fmt.Errorf("nil domain passed to BuildPolicyInput")
	}
	parent, _ := GetParentDomain(domain.ParentID)
	chain := BuildDomainChain(domain)
	return &PolicyInput{
		Self:   *domain,
		Parent: parent,
		Chain:  chain,
	}, nil
}

// BuildDomainChain traces ancestry from self → null.
func BuildDomainChain(domain *Domain) []string {
	var chain []string
	current := domain
	for current != nil {
		chain = append(chain, current.ID)
		next, _ := GetParentDomain(current.ParentID)
		current = next
	}
	return chain
}

// VerifySelfReference ensures the self pointer and chain are coherent.
func (pi *PolicyInput) VerifySelfReference() bool {
	return len(pi.Chain) > 0 && pi.Self.ID == pi.Chain[0]
}

// EvaluateSelfAwarePolicy executes a policy query using the OPA instance.
func EvaluateSelfAwarePolicy(ctx context.Context, opaInstance *rego.PreparedEvalQuery,
	domain *Domain, action interface{}) (interface{}, error) {

	input, err := BuildPolicyInput(domain)
	if err != nil {
		return nil, err
	}
	input.Action = action

	if !input.VerifySelfReference() {
		return nil, fmt.Errorf("invalid self reference for domain %s", domain.ID)
	}

	results, err := opaInstance.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("opa evaluation error: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no decision results returned")
	}
	return results[0].Expressions[0].Value, nil
}

// GetParentDomain should resolve a parent domain from DB or cache.
// Implemented elsewhere in internal/domain or internal/store.
func GetParentDomain(parentID string) (*Domain, error) {
	if parentID == "" {
		return nil, nil
	}
	// Example placeholder:
	// parent, err := store.GetDomainByID(parentID)
	// return parent, err
	return nil, nil
}
