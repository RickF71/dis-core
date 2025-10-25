package main

import (
	"dis-core/internal/ledger"
	"fmt"
	"log"
)

func main() {
	frozenHash := "15b437484377ac63cdb227b4fa264010aec06759f5808c699768cbe112f3c930"
	r := ledger.NewReceipt("domain.terra", "domain.freeze.v1", frozenHash, "ac-8d91bfa1", "uid-terracouncil-001")

	// Save it under your version folder
	saveDir := "versions/v0.6/receipts/generated"
	err := ledger.SaveReceipt(r)
	if err != nil {
		log.Fatal("❌ Failed to save receipt:", err)
	}

	fmt.Printf("✅ Receipt saved to %s/%s.json\n", saveDir, r.ReceiptID)
}
