package main

// import (
// 	"context"
// 	"flag"
// 	"fmt"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"dis-core/internal/api"
// 	"dis-core/internal/api/atlas"
// 	"dis-core/internal/config"
// 	"dis-core/internal/core"
// 	"dis-core/internal/daemon" // [v0.9.3+]
// 	"dis-core/internal/db"
// 	"dis-core/internal/policy"
// 	"dis-core/internal/version"
// )

// // 🔑 Holds verified DIS-CORE hash for runtime integrity.
// var coreHash string

// func main() {
// 	// --- CLI Flags ---
// 	cfgPath := flag.String("config", "config.yaml", "path to config file")
// 	polPath := flag.String("policy", "", "path to policy file (overrides config)")
// 	serve := flag.Bool("serve", false, "start REST API server")
// 	verifyCore := flag.Bool("verify-core", true, "verify DIS-CORE checksum at startup")
// 	flagBy := flag.String("by", "", "domain to act under (optional, headless)")
// 	flagScope := flag.String("scope", "", "scope of the act (optional, headless)")
// 	flagNonce := flag.String("nonce", "", "optional hex nonce")
// 	serve = flag.Bool("serve", false, "start DIS node server (enables API)")
// 	disPort := flag.Int("dis_port", 8080, "override config API port for this node")

// 	flag.Parse()

// 	// --- Load Config ---
// 	cfg, err := config.Load(*cfgPath)
// 	if err != nil {
// 		log.Fatal("Config error:", err)
// 	}

// 	vInfo, err := version.Load("VERSION.yaml")
// 	if err != nil {
// 		log.Printf("⚠️  Version info unavailable: %v", err)
// 	} else {
// 		log.Printf("🔐 Loaded DIS-CORE %s (frozen)", vInfo.DISCore)
// 	}

// 	// --- Verify DIS-CORE Integrity ---
// 	if *verifyCore {
// 		if hash, err := version.CoreChecksum("domains/dis/schemas/dis-core.v1.yaml"); err == nil {
// 			log.Printf("🔏 DIS-CORE integrity hash: %s", hash)
// 			coreHash = hash
// 		} else {
// 			log.Printf("⚠️  Could not verify core schema hash: %v", err)
// 		}
// 	} else {
// 		log.Println("⚠️  DIS-CORE checksum verification skipped (--verify-core=false).")
// 	}

// 	// --- Load Policy ---
// 	pPath := cfg.PolicyPath
// 	if *polPath != "" {
// 		pPath = *polPath
// 	}
// 	pol, sum, err := policy.Load(pPath)
// 	if err != nil {
// 		log.Fatal("Policy error:", err)
// 	}

// 	// --- Init / Auto-Create PostgreSQL DB (unified DIS-Core + Atlas) ---
// 	store, err := db.ConnectPostgres(cfg.DatabaseDSN)
// 	if err != nil {
// 		log.Fatalf("❌ failed to connect to PostgreSQL: %v", err)
// 	}
// 	defer store.Close()

// 	if err := db.Setup(store); err != nil {
// 		log.Fatalf("❌ failed to initialize PostgreSQL schema: %v", err)
// 	}

// 	// Initialize Atlas store wrapper (uses same Postgres connection)
// 	atlasStore := atlas.NewAtlasStore(store)

// 	// [v0.9.3+] Seed baseline identities (system + personal)
// 	_, _ = db.UpsertIdentity(store, "dis_uid:terra:system", "system", true)
// 	_, _ = db.UpsertIdentity(store, "dis_uid:terra:rick:bf72a8c19f", "rick", true)

// 	// [v0.9.3+] Context for graceful shutdown
// 	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
// 	defer cancel()

// 	// [v0.9.3+] Launch background daemon for auto-revocation
// 	go daemon.StartAutoRevocationDaemon(ctx, 60*time.Second)

// 	// --- REST Server Mode ---
// 	if *serve {
// 		port := cfg.APIPort
// 		if *disPort != 8080 {
// 			port = *disPort
// 		}
// 		addr := fmt.Sprintf("%s:%d", cfg.APIHost, port)
// 		log.Printf("🌐 DIS-NODE serving on %s (DIS-CORE %s)\n", addr, vInfo.DISCore)

// 		server := api.NewServer(store, cfg, pol, sum, coreHash)
// 		server.AttachAtlas(atlasStore)

// 		go func() {
// 			if err := server.Start(addr); err != nil {
// 				log.Fatal(err)
// 			}
// 		}()
// 		<-ctx.Done()
// 		log.Println("🛑 Shutdown signal received — closing services.")
// 		return
// 	}

// 	// --- Headless Action Mode ---
// 	if *flagBy != "" || *flagScope != "" || *flagNonce != "" {
// 		by := cfg.DefaultDomain
// 		scope := cfg.DefaultScope
// 		if *flagBy != "" {
// 			by = *flagBy
// 		}
// 		if *flagScope != "" {
// 			scope = *flagScope
// 		}

// 		recID, nonce, ts, sig, err := core.PerformConsentAction(store, by, scope, *flagNonce, cfg, pol, sum)
// 		if err != nil {
// 			log.Fatal("❌ ", err)
// 		}

// 		fmt.Printf("✅ Action recorded under %s / %s (receipt_id=%d, nonce=%s, ts=%s, sig=%s)\n",
// 			by, scope, recID, nonce, ts, sig[:16]+"...")
// 		return
// 	}

// 	// --- Default Path ---
// 	log.Println("✅ DIS-PERSONAL started successfully (PostgreSQL mode).")
// 	log.Println("Use --serve for API mode or --by/--scope for headless mode.")
// }
