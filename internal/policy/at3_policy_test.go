package policy

import (
	"context"
	"testing"
)

// Verify that the engine emits a structured deny reason when the AT-1 rule matches
func TestAT1_DenyReasonIsStructured(t *testing.T) {
	gm, err := LoadRegoBundle("gates.rego")
	if err != nil {
		t.Fatalf("load gates.rego: %v", err)
	}
	rm, err := LoadRegoBundle("risk.rego")
	if err != nil {
		t.Fatalf("load risk.rego: %v", err)
	}
	fm, err := LoadRegoBundle("freeze.rego")
	if err != nil {
		t.Fatalf("load freeze.rego: %v", err)
	}
	modules := map[string]string{"gates.rego": gm, "risk.rego": rm, "freeze.rego": fm}
	engine, err := NewEngine(modules)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	input := map[string]interface{}{
		"action": "ci.call.test.v1",
		"payload": map[string]interface{}{
			"block": true,
		},
	}

	decision, err := engine.EvaluateAction(context.Background(), input)
	if err != nil {
		t.Fatalf("evaluate action: %v", err)
	}
	if decision.Allow {
		t.Fatalf("expected deny for blocking test call; got allow")
	}
	if decision.Reason != "deny:at1.ci_call_test_block_v1" {
		t.Logf("decision details: %#v", decision.Details)
		t.Fatalf("expected structured deny reason 'deny:at1.ci_call_test_block_v1', got '%s'", decision.Reason)
	}
}
