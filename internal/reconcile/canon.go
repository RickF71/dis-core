package reconcile

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"dis-core/internal/domain"
	"dis-core/internal/ledger"
	"dis-core/internal/schema"
)

type CanonReconciler struct {
	RepoRoot string
	Schemas  *schema.Registry
	Ledger   *ledger.Ledger
}

func New(repoRoot string, schemas *schema.Registry, led *ledger.Ledger) *CanonReconciler {
	return &CanonReconciler{
		RepoRoot: repoRoot,
		Schemas:  schemas,
		Ledger:   led,
	}
}

func (r *CanonReconciler) ReconcileDomain(code string) error {
	ctx := context.Background() // ✅ add context
	code = strings.ToLower(strings.TrimSpace(code))
	if len(code) != 3 {
		return fmt.Errorf("invalid domain code %q (expect 3 letters)", code)
	}

	// 1) Load YAML
	dom, err := domain.LookupByCode(code, r.RepoRoot)
	if err != nil {
		return fmt.Errorf("lookup domain: %w", err)
	}

	// 2) Must have lineage/parent
	if len(dom.Lineage) == 0 {
		return fmt.Errorf("domain %s has no lineage defined", dom.Code)
	}
	parent := dom.Lineage[len(dom.Lineage)-1]
	if parent == "" {
		return fmt.Errorf("domain %s has no valid parent entry", dom.Code)
	}

	// 3) Ensure schema exists (name + version)
	schemaName := fmt.Sprintf("schema.%s", strings.ToLower(dom.Code))
	const v = "v1"

	if _, ok := r.Schemas.Get(schemaName, v); !ok {
		schemaName = "schema.default"
		if _, ok := r.Schemas.Get(schemaName, v); !ok {
			return fmt.Errorf("no schema found: tried %s/%s and %s/%s",
				fmt.Sprintf("schema.%s", strings.ToLower(dom.Code)), v, "schema.default", v)
		}
	}

	// 4) Record in ledger
	payload := map[string]any{
		"domain":  dom.Code,
		"name":    dom.Name,
		"parent":  parent,
		"schema":  schemaName,
		"version": v,
		"source":  filepath.Join("domains", strings.ToLower(dom.Code), fmt.Sprintf("domain.%s.yaml", strings.ToLower(dom.Code))),
		"canon":   true,
	}

	// ✅ Use RecordCall for ci.call.v1 format
	if err := r.Ledger.RecordCall(ctx, "system", dom.Code, "reconcile", "domain.reconcile.v1", payload); err != nil {
		return fmt.Errorf("ledger record: %w", err)
	}

	return nil
}
