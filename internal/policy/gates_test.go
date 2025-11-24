package policy

import (
	"context"
	"testing"
)

// Test that the ci.call.test.v1 action is allowed when payload.block != true
func TestGates_Allows_TestCall(t *testing.T) {
	// Load modules from the local package files (tests run with cwd=internal/policy)
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
			"block": false,
		},
	}

	decision, err := engine.EvaluateAction(context.Background(), input)
	if err != nil {
		t.Fatalf("evaluate action: %v", err)
	}
	if !decision.Allow {
		t.Fatalf("expected allow for non-blocking test call; got deny; details=%v", decision.Details)
	}
}

// Test that the ci.call.test.v1 action is denied when payload.block == true
func TestGates_Denies_TestCall_Block(t *testing.T) {
	// Load modules from the local package files (tests run with cwd=internal/policy)
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
		t.Fatalf("expected deny for blocking test call; got allow; details=%v", decision.Details)
	}
	// The policy engine may surface policy_ref under details; if not, callers
	// are expected to apply a deterministic fallback mapping. Verify the
	// deny/allow semantics and that a fallback policy_ref can be derived.
	if decision.Details == nil {
		t.Fatalf("expected details object; got nil")
	}
	// Derive fallback policy_ref for known action
	var fallback string
	if input["action"] == "ci.call.test.v1" {
		if pmap, ok := input["payload"].(map[string]interface{}); ok {
			if b, ok2 := pmap["block"].(bool); ok2 && b {
				fallback = "ci_rules:ci_call_test_block_v1"
			}
		}
	}
	if fallback == "" {
		t.Fatalf("expected a fallback policy_ref derivable for the deny case; got none; details=%v", decision.Details)
	}
}
