package ledger

// PolicyDecisionRecord represents a record of a policy decision made by the policy engine.
type PolicyDecisionRecord struct {
	ActorDomain  string
	AuthorizedBy string
	TargetDomain string
	Action       string
	Decision     bool
	Reason       string
	Metadata     map[string]interface{}
}
