package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"dis-core/internal/domain"
	"dis-core/internal/ledger"
	"dis-core/internal/receipts"
	"dis-core/internal/schema"
)

var (
	schemasDir    = flag.String("schemas", "schemas", "directory of schema files")
	domainsDir    = flag.String("domains", "domains", "directory of domain files")
	freezeVer     = flag.String("freeze", "", "freeze version (optional)")
	listReceipts  = flag.Bool("list-receipts", false, "list all stored receipts")
	verifyReceipt = flag.String("verify-receipt", "", "verify a specific receipt id")
)

func main() {

	flag.Parse()

	// Open ledger (auto-initializes)
	led, err := ledger.Open("data/dis_core.db")
	if err != nil {
		log.Fatalf("open ledger: %v", err)
	}

	if *listReceipts {
		list, err := led.ListReceipts()
		if err != nil {
			log.Fatalf("list receipts: %v", err)
		}
		fmt.Println("📜 Receipts:")
		for _, r := range list {
			fmt.Printf(" - %s | %s | %s | %s\n", r.ReceiptID, r.Action, r.Timestamp, r.By)
		}
		return
	}

	if *verifyReceipt != "" {
		if err := led.VerifyReceipt(*verifyReceipt); err != nil {
			log.Fatalf("verify: %v", err)
		}
		return
	}

	reg := schema.NewRegistry()
	if err := reg.LoadDir(*schemasDir); err != nil {
		log.Fatalf("load schemas: %v", err)
	}
	if err := reg.Verify("domain.notech", "v0.1"); err != nil {
		log.Printf("warn: domain.notech@v0.1 verification skipped or failed: %v", err)
	}

	if *freezeVer != "" {
		doFreeze(reg, *freezeVer)
		return
	}

	pattern := filepath.Join(*domainsDir, "*.yaml")
	files, _ := filepath.Glob(pattern)
	for _, f := range files {
		d, err := domain.LoadAndValidate(f, reg)
		if err != nil {
			log.Printf("❌ %s: %v", f, err)
			continue
		}
		fmt.Printf("✅ loaded domain: %s (%s)\n", d.Meta.Name, d.Meta.UUID)
	}
}

func doFreeze(reg *schema.Registry, version string) {
	hash := reg.HashAll()
	r := receipts.NewReceipt(
		"domain.terra",
		fmt.Sprintf("freeze_core_%s", version),
		hash,
		"console-001",
		"seat.core.architect",
	)

	dir := fmt.Sprintf("versions/%s/receipts", version)
	if err := r.Save(dir); err != nil {
		log.Fatalf("save receipt: %v", err)
	}

	led, err := ledger.Open("data/dis_core.db")
	if err == nil {
		led.InsertReceipt(r)
	}

	fmt.Printf("🔏 DIS-CORE %s frozen — hash=%s\n", version, hash[:12])
}
