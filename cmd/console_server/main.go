package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dis-core/internal/console"
	"dis-core/internal/ledger"
)

type ActionRequest struct {
	Action    string `json:"action"`
	PolicyRef string `json:"policy_ref"`
	Initiator string `json:"initiator"`
	Verify    bool   `json:"verify"` // optional flag to trigger audit
}

// --- mockResponseWriter ---
// Used for calling internal routes like /api/verify/all programmatically
type mockResponseWriter struct {
	header http.Header
	data   []byte
	status int
}

func (m *mockResponseWriter) Header() http.Header {
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	m.data = append(m.data, b...)
	return len(b), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

func main() {
	// Initialize Authority Console
	seats := []string{"uid-terracouncil-001", "uid-terracouncil-002"}
	ac := console.NewConsole("domain.terra", "DIS-CORE v1.0", seats)
	console.LoadLastVerification()

	// Load network configuration
	netCfg, err := console.LoadNetworkConfig("versions/v0.7/network.yaml")
	if err != nil {
		log.Fatalf("❌ Failed to load network config: %v", err)
	}

	// Load or create the trust ledger
	ledgerPath := "versions/v0.7/ledger/trust.json"
	trust, err := ledger.LoadTrustLedger(ledgerPath)
	if err != nil {
		log.Fatalf("❌ Failed to load trust ledger: %v", err)
	}

	// --- Background 30-minute verification loop ---
	go func() {
		for {
			log.Println("🕒 Checking if verification is needed...")

			performed, report, receipt, err := ac.RunVerificationIfNeeded()
			if !performed {
				log.Println("🕒 Scheduled verification skipped — no new receipts since last audit")
			} else if err != nil {
				log.Printf("❌ Scheduled verification failed: %v", err)
			} else {
				log.Printf("✅ Scheduled verification complete: %d valid, %d invalid — receipt %s",
					report.Valid, report.Invalid, receipt.ReceiptID)
			}

			time.Sleep(30 * time.Minute)
		}
	}()

	// --- Outbound Heartbeat Publisher ---
	// Runs every 30 minutes and pushes the latest verification receipt to trusted peers.
	go func() {
		for {
			log.Println("🌍 Heartbeat: publishing latest verification receipt to peers...")

			// Find the most recent domain.verify.v1 receipt
			dir := "versions/v0.6/receipts/generated"
			var latestFile string
			var latestTime time.Time
			filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if err == nil && !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
					if info, err := os.Stat(path); err == nil && info.ModTime().After(latestTime) {
						latestFile = path
						latestTime = info.ModTime()
					}
				}
				return nil
			})

			if latestFile == "" {
				log.Println("⚠️  No verification receipts found to publish.")
				time.Sleep(30 * time.Minute)
				continue
			}

			data, err := os.ReadFile(latestFile)
			if err != nil {
				log.Printf("❌ Failed to read receipt %s: %v", latestFile, err)
				time.Sleep(30 * time.Minute)
				continue
			}

			// Broadcast to trusted peers
			for _, peer := range netCfg.Peers {
				if peer.TrustLevel != "trusted" {
					continue
				}

				resp, err := http.Post(peer.URL+"/api/verify/external", "application/json", bytes.NewReader(data))

				entry := ledger.TrustEntry{
					Peer:       peer.Name,
					Action:     "sent",
					ReceiptID:  filepath.Base(latestFile),
					CoreHash:   "DIS-CORE v1.0",
					VerifiedAt: time.Now().UTC(),
				}

				if err != nil {
					entry.Status = "unreachable"
					entry.Notes = err.Error()
					trust.Add(entry)
					log.Printf("🌐 Failed to reach %s (%s): %v", peer.Name, peer.URL, err)
					continue
				}
				//defer resp.Body.Close()

				var result map[string]any
				_ = json.NewDecoder(resp.Body).Decode(&result)
				entry.Status = "ok"
				trust.Add(entry)

				log.Printf("🤝 Heartbeat result for %s: %v", peer.Name, result)
			}

			time.Sleep(30 * time.Minute)
		}
	}()

	// ===============================
	// === POST /api/console/action ===
	// ===============================
	http.HandleFunc("/api/console/action", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		act, err := ac.LogAction(req.Action, req.PolicyRef, req.Initiator)
		if err != nil {
			log.Printf("❌ %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ac.SaveState()

		resp := map[string]any{
			"status":  "ok",
			"action":  act.Type,
			"receipt": act.Receipt,
		}

		// Optional auto-verification trigger
		if req.Verify {
			log.Println("🔍 Auto-verification triggered...")
			reportReq, _ := http.NewRequest("GET", "/api/verify/all", nil)
			rw := &mockResponseWriter{}
			http.DefaultServeMux.ServeHTTP(rw, reportReq)

			var report map[string]any
			json.Unmarshal(rw.data, &report)
			resp["verification_report"] = report
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

		log.Printf("✅ Action %s logged from %s\n", act.Type, req.Initiator)
	})

	// ===============================
	// === GET /api/console/state ===
	// ===============================
	http.HandleFunc("/api/console/state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ac)
	})

	// ===========================
	// === GET /api/receipts ====
	// ===========================
	http.HandleFunc("/api/receipts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		dir := "versions/v0.6/receipts/generated"
		files := []string{}

		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
				files = append(files, filepath.Base(path))
			}
			return nil
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"receipts": files})
	})

	// ===================================
	// === GET /api/receipts/{id}.json ===
	// ===================================
	http.HandleFunc("/api/receipts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/api/receipts/")
		if id == "" {
			http.Error(w, "missing receipt ID", http.StatusBadRequest)
			return
		}

		file := filepath.Join("versions/v0.6/receipts/generated", id)
		if !strings.HasSuffix(file, ".json") {
			file += ".json"
		}

		data, err := os.ReadFile(file)
		if err != nil {
			http.Error(w, "receipt not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	// ======================================
	// === GET /api/verify/all (with commit)
	// ======================================
	http.HandleFunc("/api/verify/all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		report, receipt, err := ac.RunVerification()
		if err != nil {
			http.Error(w, "verification failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resp := map[string]any{
			"verified_at":             report.VerifiedAt,
			"total":                   report.Total,
			"valid":                   report.Valid,
			"invalid":                 report.Invalid,
			"results":                 report.Results,
			"verification_receipt_id": receipt.ReceiptID,
			"verification_receipt":    receipt,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// =======================================
	// === POST /api/verify/external (peers) ===
	// =======================================
	http.HandleFunc("/api/verify/external", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var incoming map[string]any
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		log.Printf("🤝 Received external verification receipt from peer")
		resp := map[string]any{
			"status": "accepted",
			"time":   time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("🌐 DIS Authority Console API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
