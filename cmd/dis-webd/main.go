// --------------------------------------------------------------------
// Freeze Core Version (records a receipt in Postgres)
// --------------------------------------------------------------------

package main

import (
	"database/sql"
	"dis-core/internal/api"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"dis-core/internal/domain"
	"dis-core/internal/ledger"
	"dis-core/internal/receipts"
	"dis-core/internal/schema"
)

var schemasDir = flag.String("schemas", "schemas", "directory of schema files")
var domainsDir = flag.String("domains", "domains", "directory of domain files")
var freezeVer = flag.String("freeze", "", "freeze version (optional)")
var listReceipts = flag.Bool("list-receipts", false, "list all stored receipts")
var verifyReceipt = flag.String("verify-receipt", "", "verify a specific receipt id")

// serveFlag removed: not used, API server always starts
var finPort = flag.Int("fin_port", 8080, "Finagler API port (UI/HTTP)")
var netPort = flag.Int("net_port", 9090, "DIS-Network peer port")
var autoNetd = flag.Bool("auto", false, "Automatically start dis-netd if not reachable")

// noBrowser removed: dis-webd does not launch browser

func main() {
	flag.Parse()

	netPingURL := fmt.Sprintf("http://localhost:%d/ping", *netPort)
	var netdCmd *os.Process = nil
	reachable := false
	log.Printf("🔎 Checking for DIS-Network at %s", netPingURL)
	resp, err := http.Get(netPingURL)
	if err == nil && resp.StatusCode == 200 {
		reachable = true
		resp.Body.Close()
		log.Printf("✅ DIS-Network already running on port %d", *netPort)
	}

	if *autoNetd && !reachable {
		log.Printf("⚙️ No dis-netd detected; auto-launching local network daemon on port %d", *netPort)
		netdProc, err := os.StartProcess("/usr/bin/env", []string{"env", "go", "run", "./cmd/dis-netd", fmt.Sprintf("--net_port=%d", *netPort)}, &os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		})
		if err != nil {
			log.Fatalf("failed to launch dis-netd: %v", err)
		}
		netdCmd = netdProc
		// Wait for /ping to respond (up to 7s, 300ms interval)
		log.Printf("⏳ Waiting for dis-netd to respond at /ping...")
		success := false
		for i := 0; i < 24; i++ { // 24 * 300ms = 7.2s
			resp, err := http.Get(netPingURL)
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				success = true
				log.Printf("✅ dis-netd is now responding at /ping")
				break
			}
			time.Sleep(300 * time.Millisecond)
		}
		if !success {
			log.Printf("❌ dis-netd did not respond at /ping after 7 seconds")
		}
	}

	// --------------------------------------------------------------------
	// Connect to PostgreSQL (use env var DIS_DB_DSN or fallback)
	// --------------------------------------------------------------------
	dsn := os.Getenv("DIS_DB_DSN")
	if dsn == "" {
		dsn = "postgres://dis_user:card567@localhost:5432/dis_core?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL ledger")

	// Initialize ledger using shared connection
	led := ledger.Open(db)

	// --------------------------------------------------------------------
	// Handle CLI flags
	// --------------------------------------------------------------------

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
		// Try alternate path if not found
		altPath := "domains/notech/schemas/domain.notech.v0.1.yaml"
		if loadErr := reg.LoadDir("domains/notech/schemas"); loadErr == nil {
			if err2 := reg.Verify("domain.notech", "v0.1"); err2 == nil {
				log.Printf("info: domain.notech@v0.1 loaded from %s", altPath)
			} else {
				log.Printf("warn: domain.notech@v0.1 verification failed: %v", err2)
			}
		} else {
			log.Printf("warn: domain.notech@v0.1 verification skipped or failed: %v", err)
		}
	}

	if *freezeVer != "" {
		doFreeze(reg, db, *freezeVer)
		return
	}

	// --------------------------------------------------------------------
	// Always start API server using internal/api router
	// --------------------------------------------------------------------
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// // Load config and policy
	// cfg, err := config.Load("config.yaml")
	// if err != nil {
	// 	log.Fatalf("config load error: %v", err)
	// }

	// pol, sum, err := policy.Load(cfg.PolicyPath)
	// if err != nil {
	// 	log.Fatalf("policy load error: %v", err)
	// }

	// coreHash := "dev"

	// Create the API server
	apiServer := api.NewServer(db)

	finMux := apiServer.Mux()

	// Global CORS middleware
	corsWrapper := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	go func() {
		addr := fmt.Sprintf(":%d", *finPort)
		log.Printf("💠 API server serving on %s", addr)
		if err := http.ListenAndServe(addr, corsWrapper(finMux)); err != nil {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	<-stop
	log.Println("🛑 API shutdown signal received.")
	if netdCmd != nil {
		log.Println("🛑 Killing auto-launched dis-netd...")
		netdCmd.Kill()
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

// openBrowser removed: dis-webd does not launch or care about the UI/browser.

func doFreeze(reg *schema.Registry, db *sql.DB, version string) {
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

	led := ledger.Open(db)
	if err := led.InsertReceipt(r); err != nil {
		log.Printf("⚠️ failed to insert freeze receipt: %v", err)
	}

	fmt.Printf("🔏 DIS-CORE %s frozen — hash=%s\n", version, hash[:12])
}
