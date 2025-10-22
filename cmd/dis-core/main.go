package main

import (
	"database/sql"
	"dis-core/internal/db"
	"dis-core/internal/domain"
	"dis-core/internal/receipts"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	_ "github.com/lib/pq"

	"dis-core/internal/api"
	"dis-core/internal/canon"
	"dis-core/internal/mirrorspin"
	"dis-core/internal/schema"

	//"dis-core/internal/overlay"

	"dis-core/internal/ledger"
)

var (
	schemasDir    = flag.String("schemas", "schemas", "directory of schema files")
	domainsDir    = flag.String("domains", "domains", "directory of domain files")
	freezeVer     = flag.String("freeze", "", "freeze version (optional)")
	listReceipts  = flag.Bool("list-receipts", false, "list all stored receipts")
	verifyReceipt = flag.String("verify-receipt", "", "verify a specific receipt id")
	serveFlag     = flag.Bool("serve", false, "start REST API server")
	finPort       = flag.Int("fin_port", 8080, "Finagler API port (UI/HTTP)")
	netPort       = flag.Int("net_port", 9090, "DIS-Network peer port")
	// disPort flag removed; use finPort and netPort only
	mode = flag.String("mode", "all", "Run mode: web | net | all")

	canonImport = flag.String("canon-import", "", "Import YAMLs into DB from directory (e.g., ./bootstrap)")
	canonFreeze = flag.Bool("canon-freeze", false, "Freeze canon import (DB is authoritative)")
)

func main() {
	flag.Parse()

	// --------------------------------------------------------------------
	// Connect to PostgreSQL
	// --------------------------------------------------------------------
	database, err := db.SetupDatabase()
	if err != nil {
		log.Fatalf("database setup failed: %v", err)
	}
	defer database.Close()
	fmt.Println("✅ Database initialized via SetupDatabase")

	led, err := ledger.Open(os.Getenv("DIS_DB_DSN"), database)
	if err != nil {
		log.Fatalf("open ledger: %v", err)
	}

	// Ensure MirrorSpin table exists
	if err := mirrorspin.EnsureMirrorEventsTable(database); err != nil {
		log.Fatalf("mirror_events table init failed: %v", err)
	}

	// --------------------------------------------------------------------
	// Canon Import & Freeze Commands
	// --------------------------------------------------------------------
	if *canonImport != "" {
		imp := &canon.CanonImporter{Ledger: led}
		if err := imp.ImportDir(*canonImport); err != nil {
			log.Fatalf("canon import failed: %v", err)
		}
		return
	}

	if *canonFreeze {
		ctrl := &canon.FreezeController{Ledger: led}
		if err := ctrl.FreezeImport(); err != nil {
			log.Fatalf("canon freeze failed: %v", err)
		}
		return
	}

	// --------------------------------------------------------------------
	// Initialize schema registry once for the entire program
	// --------------------------------------------------------------------
	reg := schema.NewRegistry()
	if err := reg.LoadDir(*schemasDir); err != nil {
		log.Fatalf("load schemas: %v", err)
	}
	if err := reg.LoadDir(filepath.Join(*schemasDir, "mirrorspin")); err != nil {
		log.Printf("⚠️ could not load MirrorSpin schemas: %v", err)
	}

	// --------------------------------------------------------------------
	// Mode handler (web / net / all)
	// --------------------------------------------------------------------
	if *mode != "all" {
		switch *mode {
		case "web":
			startWebServer(database, reg, *finPort)
		case "net":
			startNetServer(*netPort)
		default:
			fmt.Println("Usage: --mode=web|net|all")
			os.Exit(1)
		}
		return
	}

	// --------------------------------------------------------------------
	// CLI commands
	// --------------------------------------------------------------------
	if *listReceipts {
		store := ledger.NewStore(database)
		list, err := store.ListReceipts()
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
		store := ledger.NewStore(database)
		if err := store.VerifyReceipt(*verifyReceipt); err != nil {
			log.Fatalf("verify: %v", err)
		}
		return
	}

	// --------------------------------------------------------------------
	// Freeze core version
	// --------------------------------------------------------------------
	if *freezeVer != "" {
		doFreeze(led, reg, *freezeVer)
		return
	}

	// --------------------------------------------------------------------
	// Main serve mode
	// --------------------------------------------------------------------
	if *serveFlag {
		runAllServers(database, reg, *finPort, *netPort)
		return
	}

	validateDomains(reg)
}

// --------------------------------------------------------------------
// Freeze version
// --------------------------------------------------------------------
func doFreeze(led *ledger.Ledger, reg *schema.Registry, version string) {
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

	if err := led.Record("core.freeze", map[string]any{
		"version": version,
		"hash":    hash,
	}); err != nil {
		log.Printf("⚠️ failed to record freeze event: %v", err)
	}

	fmt.Printf("🔏 DIS-CORE %s frozen — hash=%s\n", version, hash[:12])
}

// --------------------------------------------------------------------
// Domain validation (fallback)
// --------------------------------------------------------------------
func validateDomains(reg *schema.Registry) {
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

// --------------------------------------------------------------------
// Web & Net server runners
// --------------------------------------------------------------------
func startWebServer(db *sql.DB, reg *schema.Registry, port int) {
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("🌐 Starting DIS-Core web server on %s\n", addr)
	srv := api.NewServer(db).WithLogger(log.Default()).WithSchemas(reg)

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

	log.Fatal(http.ListenAndServe(addr, corsWrapper(srv.Mux())))
}

func startNetServer(port int) {
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("🌐 Starting DIS network node on %s\n", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			fmt.Fprintln(c, "Hello from DIS network node")
		}(conn)
	}
}

func runAllServers(db *sql.DB, reg *schema.Registry, finPort, netPort int) {
	// disPort logic removed; ports are explicit

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	if err := canon.ExportDomains(db, "domains/_auto"); err != nil {
		log.Printf("⚠️ Canon export failed: %v", err)
	} else {
		log.Println("✅ Canonical domain export complete.")
	}

	go func() {
		log.Println("🪞 MirrorSpin engine starting...")
		mirrorspin.SpinLoop(db)
	}()

	// --------------------------------------------------------------------
	// Initialize managers (needed for YAML import API)
	// --------------------------------------------------------------------
	domainMgr := domain.NewManager(db)
	schemaMgr := schema.NewManager(db)
	// policyMgr := policy.NewManager(db)  // TODO: implement policy.NewManager
	// overlay optional; uncomment later
	// overlayMgr := overlay.NewManager(db)

	// Create the API server and inject managers
	srv := api.NewServer(db).
		WithLogger(log.Default()).
		WithSchemas(reg)

	srv.DomainManager = domainMgr
	srv.SchemaManager = schemaMgr
	// srv.PolicyManager = policyMgr  // TODO: uncomment when policy.NewManager is implemented
	// srv.OverlayManager = overlayMgr

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
		addr := fmt.Sprintf(":%d", finPort)
		log.Printf("🌐 Finagler API serving on %s", addr)
		if err := http.ListenAndServe(addr, corsWrapper(srv.Mux())); err != nil {
			log.Fatalf("Finagler API server failed: %v", err)
		}
	}()

	go func() {
		addr := fmt.Sprintf(":%d", netPort)
		log.Printf("🌐 DIS-Network serving on %s", addr)
		if err := http.ListenAndServe(addr, http.NewServeMux()); err != nil {
			log.Fatalf("DIS-Network server failed: %v", err)
		}
	}()

	<-stop
	log.Println("🛑 Shutdown signal received — closing services.")
}
