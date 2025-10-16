package tests

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"dis-core/internal/db"
	"dis-core/internal/ledger"
	"dis-core/internal/types"
)

// helper: open temporary in-memory DB
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	if err := db.EnsureValueReceiptsSchema(conn); err != nil {
		t.Fatalf("failed to ensure value_receipts schema: %v", err)
	}
	if err := db.EnsureLifePushSubstrateSchema(conn); err != nil {
		t.Fatalf("failed to ensure lifepush_substrate schema: %v", err)
	}
	return conn
}

func TestSchemaRegistryIntegration(t *testing.T) {
	reg := &ledger.SchemaRegistry{}
	reg.Init() // if you have a separate Init(), otherwise initialize manually
	pathVR := "./schemas/value_receipt.v1.yaml"
	pathLS := "./schemas/lifepush_substrate_structure.v1.yaml"

	dataVR, err := os.ReadFile(filepath.Clean(pathVR))
	if err != nil {
		t.Skipf("Skipping: missing %s", pathVR)
	}
	dataLS, err := os.ReadFile(filepath.Clean(pathLS))
	if err != nil {
		t.Skipf("Skipping: missing %s", pathLS)
	}

	rec1, err := reg.RegisterSchema("value_receipt", "v1", dataVR, "system", pathVR)
	if err != nil {
		t.Fatalf("RegisterSchema value_receipt: %v", err)
	}
	rec2, err := reg.RegisterSchema("lifepush_substrate_structure", "v1", dataLS, "system", pathLS)
	if err != nil {
		t.Fatalf("RegisterSchema lifepush_substrate: %v", err)
	}

	if rec1.Hash == "" || rec2.Hash == "" {
		t.Error("expected non-empty schema hashes")
	}
}

func TestValueReceiptAndSubstrateInsert(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	// create substrate record
	sub := types.LifePushSubstrate{
		ID:                  "sub-terra-0001",
		Layer:               "Terra",
		CoherenceThreshold:  0.6,
		EnergyFlowMin:       0.4,
		ConsentIntegrityMin: 0.5,
		SuccessorLayer:      "Limen",
		ObserverDomain:      "domain.terra",
		Active:              true,
	}
	_, err := conn.Exec(`
		INSERT INTO lifepush_substrate_structure 
		(id, layer, coherence_threshold, energy_flow_min, consent_integrity_min, successor_layer, observer_domain, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sub.ID, sub.Layer, sub.CoherenceThreshold, sub.EnergyFlowMin, sub.ConsentIntegrityMin,
		sub.SuccessorLayer, sub.ObserverDomain, 1)
	if err != nil {
		t.Fatalf("failed to insert substrate: %v", err)
	}

	// create value receipt
	vr := types.ValueReceipt{
		ID:             "rcpt-value-001",
		By:             "domain.terra.usa",
		ActionRef:      "action.seed_001",
		SubstrateRef:   sub.ID,
		CoherenceDelta: 0.42,
		ValueVector: map[string]float64{
			"ethical":    0.6,
			"ecological": 0.3,
			"technical":  -0.1,
		},
		ObserverField: "domain.terra",
		Timestamp:     time.Now().UTC(),
	}
	jsonVec, _ := json.Marshal(vr.ValueVector)
	_, err = conn.Exec(`
		INSERT INTO value_receipts
		(id, by, action_ref, substrate_ref, coherence_delta, value_vector, observer_field, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		vr.ID, vr.By, vr.ActionRef, vr.SubstrateRef, vr.CoherenceDelta, string(jsonVec), vr.ObserverField, vr.Timestamp)
	if err != nil {
		t.Fatalf("failed to insert value_receipt: %v", err)
	}

	// verify round-trip
	var gotDelta float64
	err = conn.QueryRow(`SELECT coherence_delta FROM value_receipts WHERE id = ?`, vr.ID).Scan(&gotDelta)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if gotDelta != vr.CoherenceDelta {
		t.Errorf("expected %f, got %f", vr.CoherenceDelta, gotDelta)
	}
}

func TestSchemaHashStability(t *testing.T) {
	path := "./schemas/value_receipt.v1.yaml"
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Skipf("missing schema: %s", path)
	}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if len(got) != 64 {
		t.Errorf("invalid hash length: %d", len(got))
	}
	t.Logf("schema %s hash=%s", path, got[:16])
}
