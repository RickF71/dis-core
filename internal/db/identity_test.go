package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIdentitiesLifecycle(t *testing.T) {
	// Create a temp DB file under /tmp or local workspace
	tmpDir := os.TempDir()
	dbPath := filepath.Join(tmpDir, "dis_test_identities.db")

	// Clean up any old test db
	_ = os.Remove(dbPath)

	db, err := SetupDatabase(dbPath)
	if err != nil {
		t.Fatalf("failed to setup database: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Verify identities table exists
	if err := EnsureIdentitiesSchema(db); err != nil {
		t.Fatalf("EnsureIdentitiesSchema failed: %v", err)
	}

	// Insert a test identity
	disUID := "dis_uid:terra:testuser:abcd1234"
	ns := "testuser"
	id, err := InsertIdentity(db, disUID, ns)
	if err != nil {
		t.Fatalf("InsertIdentity failed: %v", err)
	}
	if id == 0 {
		t.Fatal("InsertIdentity returned 0 ID")
	}

	// Retrieve the inserted record
	found, err := GetIdentity(db, disUID)
	if err != nil {
		t.Fatalf("GetIdentity failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected identity, got nil")
	}
	if found.DISUID != disUID {
		t.Errorf("DISUID mismatch: want %s, got %s", disUID, found.DISUID)
	}
	if found.Namespace != ns {
		t.Errorf("Namespace mismatch: want %s, got %s", ns, found.Namespace)
	}

	// Deactivate the identity and verify it
	if err := DeactivateIdentity(db, disUID); err != nil {
		t.Fatalf("DeactivateIdentity failed: %v", err)
	}
	inactive, err := GetIdentity(db, disUID)
	if err != nil {
		t.Fatalf("GetIdentity after deactivate failed: %v", err)
	}
	if inactive.Active {
		t.Errorf("expected identity to be inactive after deactivation")
	}

	t.Logf("✅ Identity lifecycle test completed successfully")
}
