package main

import (
	"flag"
	"fmt"
	"log"

	"dis-core/internal/api"
	"dis-core/internal/config"
	"dis-core/internal/core"
	"dis-core/internal/db"
	"dis-core/internal/policy"
	"dis-core/internal/version"
)

// 🔑 Holds verified DIS-CORE hash for runtime integrity.
var coreHash string

func main() {
	// --- CLI Flags ---
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	polPath := flag.String("policy", "", "path to policy file (overrides config)")
	serve := flag.Bool("serve", false, "start REST API server")
	verifyCore := flag.Bool("verify-core", true, "verify DIS-CORE checksum at startup")
	flagBy := flag.String("by", "", "domain to act under (optional, headless)")
	flagScope := flag.String("scope", "", "scope of the act (optional, headless)")
	flagNonce := flag.String("nonce", "", "optional hex nonce")
	flag.Parse()

	// --- Load Config ---
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal("Config error:", err)
	}

	vInfo, err := version.Load("VERSION.yaml")
	if err != nil {
		log.Printf("⚠️  Version info unavailable: %v", err)
	} else {
		log.Printf("🔐 Loaded DIS-CORE %s (frozen)", vInfo.DISCore)
	}

	// --- Verify DIS-CORE Integrity ---
	if *verifyCore {
		if hash, err := version.CoreChecksum("schemas/dis-core.v1.yaml"); err == nil {
			log.Printf("🔏 DIS-CORE integrity hash: %s", hash)
			coreHash = hash
		} else {
			log.Printf("⚠️  Could not verify core schema hash: %v", err)
		}
	} else {
		log.Println("⚠️  DIS-CORE checksum verification skipped (--verify-core=false).")
	}

	// --- Load Policy ---
	pPath := cfg.PolicyPath
	if *polPath != "" {
		pPath = *polPath
	}
	pol, sum, err := policy.Load(pPath)
	if err != nil {
		log.Fatal("Policy error:", err)
	}

	// --- Init / Auto-Create DB ---
	store, err := db.SetupDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatal("DB setup failed:", err)
	}
	defer store.Close()

	// --- REST Server Mode ---
	if *serve {
		addr := fmt.Sprintf("%s:%d", cfg.APIHost, cfg.APIPort)
		log.Printf("🌐 DIS-PERSONAL %s — Network Sovereignty (bound to DIS-CORE %s) serving on %s\n",
			vInfo.DISPersonal, vInfo.DISCore, addr)

		server := api.NewServer(store, cfg, pol, sum, coreHash)
		if err := server.Start(addr); err != nil {
			log.Fatal(err)
		}
		return
	}

	// dbh, _ := db.InitDB("data/dis.db")
	// dbh.Exec(`INSERT INTO identities (id) VALUES ('demo-uid-1')`)
	// fmt.Println("✅ Created demo identity: demo-uid-1")

	// --- Headless Action Mode ---
	if *flagBy != "" || *flagScope != "" || *flagNonce != "" {
		by := cfg.DefaultDomain
		scope := cfg.DefaultScope
		if *flagBy != "" {
			by = *flagBy
		}
		if *flagScope != "" {
			scope = *flagScope
		}

		recID, nonce, ts, sig, err := core.PerformConsentAction(store, by, scope, *flagNonce, cfg, pol, sum)
		if err != nil {
			log.Fatal("❌ ", err)
		}

		fmt.Printf("✅ Action recorded under %s / %s (receipt_id=%d, nonce=%s, ts=%s, sig=%s)\n",
			by, scope, recID, nonce, ts, sig[:16]+"...")
		return
	}

	// --- Default Path ---
	log.Println("✅ DIS-PERSONAL started successfully (no interactive console).")
	log.Println("Use --serve for API mode or --by/--scope for headless mode.")
}
