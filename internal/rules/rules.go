package rules

import (
	"dis-core/internal/events"
	"log"
	"strings"
)

// Action represents the result of interpreting an event.
type Action struct {
	Type        string         `json:"type"`
	TrustDelta  float64        `json:"trust_delta"`
	EthicsDelta float64        `json:"ethics_delta"`
	Receipt     bool           `json:"receipt"`
	Context     map[string]any `json:"context,omitempty"`

	// --- v0.8.8 optional lineage fields ---
	ConsentRef  string `json:"consent_ref,omitempty"`  // links to ConsentGate.ID
	FeedbackRef string `json:"feedback_ref,omitempty"` // links to TrustFeedback.FeedbackID
}

// BehaviorRule defines one rule in the behavior set.
type BehaviorRule struct {
	ID        string
	EventType string
	Condition string
	Action    Action
}

// BehaviorSet holds all rules loaded from YAML.
type BehaviorSet struct {
	Rules []BehaviorRule
}

// Decide interprets an incoming event using matching rules.
func (bs *BehaviorSet) Decide(e events.Event) Action {
	for _, r := range bs.Rules {
		if r.EventType == e.Type && matchCondition(r.Condition, e.Context) {
			log.Printf("🧩 Rule matched: %s (%s)", r.ID, e.Type)
			return r.Action
		}
	}
	// Default: no-op action
	return Action{Type: "none", TrustDelta: 0.0, EthicsDelta: 0.0, Receipt: false}
}

// matchCondition performs a simple string equality check for now.
func matchCondition(cond string, ctx map[string]any) bool {
	if cond == "" {
		return true
	}
	// Basic example: condition format "key == value"
	parts := strings.Split(cond, "==")
	if len(parts) != 2 {
		return false
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if ctx[key] == val {
		return true
	}
	return false
}
