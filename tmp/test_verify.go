package main

import (
	"fmt"
	"os"

	"dis-core/internal/schema"
)

func main() {
	// Pretend we already loaded registry entries
	schemas := []schema.Entry{
		{
			ID:      "dis-core.v1.yaml",
			Version: "0.9.2",
			Hash:    "placeholder", // will be computed later
			Path:    "schemas/dis-core.v1.yaml",
		},
		{
			ID:      "identity.v0.yaml",
			Version: "0.9.2",
			Hash:    "placeholder",
			Path:    "schemas/identity.v0.yaml",
		},
	}

	if err := schema.VerifySchemaSet(schemas); err != nil {
		fmt.Println("❌ Verification failed:", err)
		os.Exit(1)
	}

	fmt.Println("✅ All schemas verified successfully")
}
