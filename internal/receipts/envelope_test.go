package receipts

import (
	"testing"
)

func TestNewEnvelope_Basic(t *testing.T) {
	originID := "dom-1"
	originName := "example"
	env := NewEnvelope(originID, originName, "actor-1")
	if env.DomainID != "dom-1" {
		t.Fatalf("expected DomainID dom-1, got %s", env.DomainID)
	}
	if env.ActorID != "actor-1" {
		t.Fatalf("expected ActorID actor-1, got %s", env.ActorID)
	}
	if env.ActionPanel == nil || env.PolicyPanel == nil || env.DomainPanel == nil {
		t.Fatalf("expected panels to be initialized")
	}
}
