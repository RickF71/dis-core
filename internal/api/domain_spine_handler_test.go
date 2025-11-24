package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dis-core/internal/core/domain/spineconfig"
	"dis-core/internal/ledger"
	testdb "dis-core/internal/testdb"
)

func TestHandleDomainSpine(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer func() { pool.Close() }()

	// Seed full canonical spine
	if err := testdb.SeedCanonicalSpine(pool); err != nil {
		t.Fatalf("failed to seed canonical spine: %v", err)
	}

	// Open ledger backed by the test pool
	led, err := ledger.Open(context.Background(), "", pool, nil)
	if err != nil {
		t.Fatalf("failed to open ledger: %v", err)
	}
	defer led.Close()

	// Build API server (policy engine and authority engine not required for this handler)
	s := NewWithPolicy(pool, led, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/domain/spine", nil)
	rr := httptest.NewRecorder()

	s.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	// Handler returns a top-level JSON array
	var resp []struct {
		Name      string  `json:"name"`
		ID        string  `json:"id"`
		ParentID  *string `json:"parent_id"`
		Dimension int     `json:"dimension"`
		PSeat     struct {
			Occupied bool `json:"occupied"`
		} `json:"pseat"`
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad JSON: %v; body=%s", err, rr.Body.String())
	}

	spine := spineconfig.CanonicalSpine()

	if len(resp) != len(spine) {
		t.Fatalf("spine length mismatch: expected %d, got %d", len(spine), len(resp))
	}

	for i, entry := range spine {
		got := resp[i]

		if got.Name != entry.Name {
			t.Errorf("spine[%d].name: expected %q, got %q", i, entry.Name, got.Name)
		}

		if got.Dimension != entry.Dimension {
			t.Errorf("spine[%d].dimension: expected %d, got %d", i, entry.Dimension, got.Dimension)
		}

		if got.PSeat.Occupied != false {
			t.Errorf("spine[%d].pseat.occupied: expected false, got true", i)
		}

		_, hasParent := spineconfig.ParentName(entry.Name)
		if !hasParent {
			if got.ParentID != nil {
				t.Errorf("spine[%d].parentId: expected nil (root), got %v", i, *got.ParentID)
			}
		} else {
			if got.ParentID == nil {
				t.Errorf("spine[%d].parentId: expected non-nil parent for %q", i, entry.Name)
			}
		}

		if got.ID == "" {
			t.Errorf("spine[%d].id is empty UUID", i)
		}
	}
}
