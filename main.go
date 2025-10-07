package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"dis-core/internal/api"
	"dis-core/internal/config"
	"dis-core/internal/core"
	"dis-core/internal/db"
	"dis-core/internal/policy"
	"dis-core/internal/version"
)

// 🔑 Holds the verified Core hash for use across runtime.
var coreHash string

func main() {
	// Flags for headless or server invocation
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	polPath := flag.String("policy", "", "path to policy file (overrides config)")
	serve := flag.Bool("serve", false, "start REST API server")
	verifyCore := flag.Bool("verify-core", true, "Verify DIS-CORE schema checksum at startup (recommended)")
	flagBy := flag.String("by", "", "domain to act under (overrides config; headless)")
	flagScope := flag.String("scope", "", "scope of the act (overrides config; headless)")
	flagNonce := flag.String("nonce", "", "hex nonce to use (optional; headless)")
	flag.Parse()

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

	// 🔍 Checksum verification toggle
	if *verifyCore {
		if hash, err := version.CoreChecksum("schemas/dis-core.v1.yaml"); err == nil {
			log.Printf("🔏 DIS-CORE integrity hash: %s", hash)
			coreHash = hash
		} else {
			log.Printf("⚠️  Could not verify core schema hash: %v", err)
		}
	} else {
		log.Println("⚠️  DIS-CORE checksum verification skipped (use --verify-core to enable).")
	}

	pPath := cfg.PolicyPath
	if *polPath != "" {
		pPath = *polPath
	}
	pol, sum, err := policy.Load(pPath)
	if err != nil {
		log.Fatal("Policy error:", err)
	}

	store, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatal("DB init failed:", err)
	}
	defer store.Close()

	if *serve {
		s := api.NewServer(store, cfg, pol, sum, coreHash)
		addr := fmt.Sprintf("%s:%d", cfg.APIHost, cfg.APIPort)
		log.Printf("🌐 DIS-PERSONAL %s — Network Sovereignty (bound to DIS-CORE %s) serving on %s\n", vInfo.DISPersonal, vInfo.DISCore, addr)
		if err := s.Start(addr); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Headless act path
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
		fmt.Printf("✅ Action recorded under %s / %s (receipt_id=%d, nonce=%s, ts=%s, sig=%s)\n", by, scope, recID, nonce, ts, sig[:16]+"...")
		return
	}

	// Interactive console
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n🌐 Direct Individual Sovereignty v0.5 — Network Sovereignty")
		fmt.Println("1) Create new identity")
		fmt.Println("2) Perform contextual action (policy-validated)")
		fmt.Println("3) View receipts")
		fmt.Println("4) Reset database (self-healing)")
		fmt.Println("5) Show config & policy")
		fmt.Println("S) Start REST server (Ctrl+C to stop)")
		fmt.Println("X) Exit")
		fmt.Print("> ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToUpper(choice))

		switch choice {
		case "1":
			uid := core.NewIdentity(store)
			fmt.Println("✅ New DIS_UID:", uid)

		case "2":
			by := cfg.DefaultDomain
			scope := cfg.DefaultScope

			fmt.Printf("By domain [%s]: ", by)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				by = input
			}

			fmt.Printf("Scope [%s]: ", scope)
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				scope = input
			}

			recID, nonce, ts, sig, err := core.PerformConsentAction(store, by, scope, "", cfg, pol, sum)
			if err != nil {
				log.Println("❌", err)
				break
			}
			fmt.Printf("🪶 Consent recorded (id=%d, nonce=%s, ts=%s, sig=%s)\n", recID, nonce, ts, sig[:16]+"...")

		case "3":
			core.ListReceipts(store)

		case "4":
			os.Remove(cfg.DatabasePath)
			if _, err := db.InitDB(cfg.DatabasePath); err != nil {
				log.Println("Reset failed to re-init:", err)
			} else {
				fmt.Println("🧹 Database erased & re-initialized.")
			}

		case "5":
			fmt.Printf("  domain: %s\n  scope:  %s\n  db:     %s\n  policy: %s\n  api:    %s:%d\n",
				cfg.DefaultDomain, cfg.DefaultScope, cfg.DatabasePath, pPath, cfg.APIHost, cfg.APIPort)
			fmt.Printf("Policy checksum: %s\n", sum)
			policy.PrintSummary(pol)

		case "S":
			s := api.NewServer(store, cfg, pol, sum, coreHash)
			addr := fmt.Sprintf("%s:%d", cfg.APIHost, cfg.APIPort)
			log.Printf("🌐 Serving on http://%s (Ctrl+C to stop)\n", addr)
			if err := s.Start(addr); err != nil {
				log.Println(err)
			}

		case "X":
			fmt.Println("👋 Exiting DIS-PERSONAL v0.5. Goodbye.")
			return

		default:
			fmt.Println("Invalid option. Please choose 1–5, S, or X to exit.")
		}
	}
}
