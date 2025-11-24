package policy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/opa/rego"
)

// Eval evaluates a specific policy package by name (e.g., "dis.policy.at1_example").
// It returns (allow, details, error). Details may be nil.
func Eval(ctx context.Context, pkg string, input map[string]interface{}) (bool, map[string]interface{}, error) {
	// Map package name to expected file path under internal/policy
	// e.g., dis.policy.at1_example -> rules/at1_example.rego
	// naive mapping: take last part after '.' and use as filename
	parts := filepath.Base(pkg)
	// parts will be like "at1_example" for our package
	fileName := fmt.Sprintf("rules/%s.rego", parts)
	path := filepath.Join("internal", "policy", fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return false, nil, fmt.Errorf("policy eval: read %s: %w", path, err)
	}

	query := fmt.Sprintf("data.%s.allow", pkg)
	q := rego.New(
		rego.Query(query),
		rego.Module(fileName, string(data)),
	)
	prepared, err := q.PrepareForEval(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("policy eval: prepare: %w", err)
	}
	rs, err := prepared.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, nil, fmt.Errorf("policy eval: eval: %w", err)
	}
	allow := false
	if len(rs) > 0 && len(rs[0].Expressions) > 0 {
		if b, ok := rs[0].Expressions[0].Value.(bool); ok {
			allow = b
		}
	}
	// Return basic details: include allow under key pkg
	details := map[string]interface{}{fmt.Sprintf("%s.allow", parts): allow}
	return allow, details, nil
}
