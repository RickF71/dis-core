package main

import (
	"database/sql"
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
	"dis-core/internal/domain"
	"dis-core/internal/ledger"
	"dis-core/internal/mirrorspin"
	"dis-core/internal/receipts"
	"dis-core/internal/schema"
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
	disPort       = flag.Int("dis_port", 0, "legacy: overrides net_port if provided")
	mode          = flag.String("mode", "all", "Run mode: web | net | all")
)

func main() {
	flag.Parse()

	// --------------------------------------------------------------------
	// Connect to PostgreSQL
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
	fmt.Println("✅ Connected to PostgreSQL ledger")

	led := ledger.Open(db)

	// Ensure MirrorSpin table exists
	if err := mirrorspin.EnsureMirrorEventsTable(db); err != nil {
		log.Fatalf("mirror_events table init failed: %v", err)
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
			startWebServer(db, reg)
		case "net":
			startNetServer()
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

	// --------------------------------------------------------------------
	// Freeze core version
	// --------------------------------------------------------------------
	if *freezeVer != "" {
		doFreeze(reg, db, *freezeVer)
		return
	}

	// --------------------------------------------------------------------
	// Main serve mode
	// --------------------------------------------------------------------
	if *serveFlag {
		runAllServers(db, reg)
		return
	}

	// Fallback: validate all domain files
	validateDomains(reg)
}

// --------------------------------------------------------------------
// Web & Net server runners
// --------------------------------------------------------------------
func startWebServer(db *sql.DB, reg *schema.Registry) {
	fmt.Printf("🌐 Starting DIS-Core web server on port 8080\n")
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
	log.Fatal(http.ListenAndServe(":8080", corsWrapper(srv.Mux())))
}

func startNetServer() {
	fmt.Printf("🌐 Starting DIS network node on port 9090\n")
	listener, err := net.Listen("tcp", ":9090")
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

func runAllServers(db *sql.DB, reg *schema.Registry) {
	finPortVal := *finPort
	netPortVal := *netPort

	if *disPort != 0 {
		netPortVal = *disPort
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Canonize on startup
	if err := canon.ExportDomains(db, "domains/_auto"); err != nil {
		log.Printf("⚠️ Canon export failed: %v", err)
	} else {
		log.Println("✅ Canonical domain export complete.")
	}

	// 🪞 Start MirrorSpin reflection loop
	go func() {
		log.Println("🪞 MirrorSpin engine starting...")
		mirrorspin.SpinLoop(db)
	}()

	// Build API server
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

	// Start Finagler (web)
	go func() {
		addr := fmt.Sprintf(":%d", finPortVal)
		log.Printf("🌐 Finagler API serving on %s", addr)
		if err := http.ListenAndServe(addr, corsWrapper(srv.Mux())); err != nil {
			log.Fatalf("Finagler API server failed: %v", err)
		}
	}()

	// Start network
	go func() {
		addr := fmt.Sprintf(":%d", netPortVal)
		log.Printf("🌐 DIS-Network serving on %s", addr)
		if err := http.ListenAndServe(addr, http.NewServeMux()); err != nil {
			log.Fatalf("DIS-Network server failed: %v", err)
		}
	}()

	<-stop
	log.Println("🛑 Shutdown signal received — closing services.")
}

// --------------------------------------------------------------------
// Freeze version
// --------------------------------------------------------------------
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
