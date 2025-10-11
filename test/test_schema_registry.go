package main

import (
	"dis-core/internal/ledger"
	"fmt"
	"path/filepath"
)

func main() {
	// Point this to your project root that contains /schemas
	root := "."
	version := "v0.8.7"

	ld, err := ledger.NewLedger(root, version)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Ledger created.")
	fmt.Println(ld.DumpSchemas())

	// Optional: manually register one more schema to verify runtime registration
	data := []byte("version: v0.99\nid: test_dynamic\n")
	rec, err := ld.RegisterSchema("test_dynamic", "v0.99", data, "test", filepath.Join(root, "schemas/test_dynamic.yaml"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("🧩 Registered dynamically: %+v\n", rec)
	fmt.Println("🧾 Registry after update:")
	fmt.Println(ld.DumpSchemas())
}
