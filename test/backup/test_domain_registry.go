//go:build ignore

package main

import (
	"dis-core/internal/ledger"
	"fmt"
	"path/filepath"
)

// test_domain_registry.go
// Simple validation and introspection test for domain-schema binding.

func main() {
	root := "." // assume running from repo root
	version := "v0.8.7"

	fmt.Printf("🧩 DIS Domain Registry Test — %s\n", version)
	fmt.Println("🔍 Initializing ledger and loading schemas...")

	ld, err := ledger.NewLedger(root, version)
	if err != nil {
		panic(fmt.Errorf("failed to init ledger: %v", err))
	}

	// Domain path
	domainPath := filepath.Join(root, "domains")
	fmt.Printf("🌐 Scanning domain path: %s\n", domainPath)

	domains, err := ld.LoadDomainsFromFS(domainPath)
	if err != nil {
		panic(fmt.Errorf("domain load error: %v", err))
	}

	valid := 0
	for _, d := range domains {
		status := "❌ invalid"
		if d.Validated && d.IsBound {
			status = "✅ valid"
			valid++
		}
		fmt.Printf("[%s] %s — schema_ref: %s (%s)\n", status, d.ID, d.SchemaRef, d.Version)
	}

	fmt.Printf("\nSummary: %d/%d valid domains\n", valid, len(domains))
	fmt.Println("---------------------------------------")

	for _, d := range domains {
		fmt.Printf("ID: %s\nSchema: %s\nVersion: %s\nValidated: %v\nBound: %v\nSource: %s\nCheckedAt: %v\n",
			d.ID, d.SchemaRef, d.Version, d.Validated, d.IsBound, d.SourcePath, d.CheckedAt)
		fmt.Println("---------------------------------------")
	}
}
