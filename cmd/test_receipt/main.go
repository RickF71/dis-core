package main

import (
	"dis-core/internal/ledger"
	"fmt"
	"log"
)

func main() {
	frozenHash := "15b437484377ac63cdb227b4fa264010aec06759f5808c699768cbe112f3c930"
	r, err := ledger.NewReceipt("domain.terra", "domain.freeze.v1", frozenHash, "ac-8d91bfa1", "uid-terracouncil-001")
	if err != nil {
		log.Fatal("❌ Failed to create receipt:", err)
	}

	// TODO: Add database connection to save receipt
	// ctx := context.Background()
	// saveErr := r.Save(ctx, conn)
	// if saveErr != nil {
	// 	log.Fatal("❌ Failed to save receipt:", saveErr)
	// }

	fmt.Printf("✅ Receipt created with ID: %s\n", r.ID)
}
