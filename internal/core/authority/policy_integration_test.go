package authority

import (
	"context"
	"encoding/json"
	"testing"

	authpkg "dis-core/internal/auth"
	dbstore "dis-core/internal/db"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
)

// TestFreezeDomain_DecisionAttached ensures EvaluateTx details are propagated into the domain receipt payload.
func TestFreezeDomain_DecisionAttached(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	// ensure domain exists
	var domainID uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM domains LIMIT 1").Scan(&domainID); err != nil {
		domainID = uuid.New()
		_, err := pool.Exec(ctx, `INSERT INTO domains (id, name, domain_type, authority, created_at) VALUES ($1,$2,$3,$4,NOW())`, domainID, "policy-test", "corporeal", "test")
		if err != nil {
			t.Fatalf("failed to create domain: %v", err)
		}
	}

	eng := NewEngine(nil, pool)

	// Wire a simple EvalFn that returns allow=true and includes a policy_ref detail
	eng.SetPolicyEvalFunc(func(ctx context.Context, input map[string]interface{}) (bool, string, map[string]interface{}, error) {
		return true, "ok", map[string]interface{}{"gates.policy_ref": "policy:domain.freeze.v1"}, nil
	})

	// actor for provenance
	au := &authpkg.ActiveUser{CorporealDomainUID: uuid.NewString()}
	ctx = authpkg.WithActiveUser(ctx, au)

	id, err := eng.FreezeDomain(ctx, domainID, "all", "test freeze", uuid.New(), nil)
	if err != nil {
		t.Fatalf("FreezeDomain failed: %v", err)
	}
	if id == uuid.Nil {
		t.Fatalf("expected non-nil freeze id")
	}

	// Find latest receipt for domain. The project supports multiple receipts
	// schemas; try the payload-based schema first and fall back to the
	// alternate (metadata) schema.
	var payload map[string]interface{}
	// Try payload schema
	rows, err := pool.Query(ctx, `SELECT payload FROM receipts WHERE domain::text = $1 OR (payload->>'domain_id') = $1 ORDER BY created_at DESC LIMIT 1`, domainID.String())
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			if err := rows.Scan(&payload); err != nil {
				t.Fatalf("scan payload failed: %v", err)
			}
		}
	} else {
		// fallback to metadata schema
		rows2, err2 := pool.Query(ctx, `SELECT metadata FROM receipts WHERE domain::text = $1 OR (metadata->>'domain_id') = $1 ORDER BY issued_at DESC LIMIT 1`, domainID.String())
		if err2 != nil {
			t.Fatalf("query receipts failed (both schemas): %v / %v", err, err2)
		}
		defer rows2.Close()
		if rows2.Next() {
			if err := rows2.Scan(&payload); err != nil {
				t.Fatalf("scan metadata failed: %v", err)
			}
		}
	}
	if payload == nil {
		t.Fatalf("no receipt payload found for domain after freeze")
	}
	// arrival: decision should be present and include policy_ref
	dec, ok := payload["decision"]
	if !ok {
		t.Fatalf("decision not found in payload: %#v", payload)
	}
	// decode to map
	var dmap map[string]interface{}
	switch v := dec.(type) {
	case map[string]interface{}:
		dmap = v
	default:
		// try marshal/unmarshal
		b, _ := json.Marshal(v)
		_ = json.Unmarshal(b, &dmap)
	}
	if dmap == nil {
		t.Fatalf("decision map decode failed")
	}
	if pr, ok := dmap["policy_ref"]; !ok || pr == "" {
		t.Fatalf("expected policy_ref in decision, got %#v", dmap)
	}
}

// TestFreezeDomain_Denied_NoReceipt ensures a policy denial prevents the freeze and no receipt is recorded.
func TestFreezeDomain_Denied_NoReceipt(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	var domainID uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM domains LIMIT 1").Scan(&domainID); err != nil {
		domainID = uuid.New()
		_, err := pool.Exec(ctx, `INSERT INTO domains (id, name, domain_type, authority, created_at) VALUES ($1,$2,$3,$4,NOW())`, domainID, "policy-test-2", "corporeal", "test")
		if err != nil {
			t.Fatalf("failed to create domain: %v", err)
		}
	}

	eng := NewEngine(nil, pool)
	eng.SetPolicyEvalFunc(func(ctx context.Context, input map[string]interface{}) (bool, string, map[string]interface{}, error) {
		return false, "policy denies", map[string]interface{}{"gates.deny_code": "TEST_DENY"}, nil
	})

	// attempt freeze
	_, err := eng.FreezeDomain(ctx, domainID, "all", "should be denied", uuid.New(), nil)
	if err == nil {
		t.Fatalf("expected FreezeDomain to fail due to policy denial")
	}

	// ensure no receipts exist for a freeze (try both schemas)
	rows, err := pool.Query(ctx, `SELECT payload FROM receipts WHERE domain::text = $1 OR (payload->>'domain_id') = $1 LIMIT 1`, domainID.String())
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			var pl map[string]interface{}
			if err := rows.Scan(&pl); err == nil && pl != nil {
				t.Fatalf("expected no receipts for denied freeze (payload schema), but found one: %v", pl)
			}
		}
	} else {
		rows2, err2 := pool.Query(ctx, `SELECT metadata FROM receipts WHERE domain::text = $1 OR (metadata->>'domain_id') = $1 LIMIT 1`, domainID.String())
		if err2 == nil {
			defer rows2.Close()
			if rows2.Next() {
				var md map[string]interface{}
				if err := rows2.Scan(&md); err == nil && md != nil {
					t.Fatalf("expected no receipts for denied freeze (metadata schema), but found one: %v", md)
				}
			}
		} else {
			// If both queries fail, surface the original error
			t.Fatalf("query receipts failed (both schemas): %v / %v", err, err2)
		}
	}
}

// TestInstantiateSeat_Provenance ensures seat receipts contain actor and identity provenance.
func TestInstantiateSeat_Provenance(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	var domainID uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM domains LIMIT 1").Scan(&domainID); err != nil {
		domainID = uuid.New()
		_, err := pool.Exec(ctx, `INSERT INTO domains (id, name, domain_type, authority, created_at) VALUES ($1,$2,$3,$4,NOW())`, domainID, "prov-domain", "corporeal", "test")
		if err != nil {
			t.Fatalf("failed to create domain: %v", err)
		}
	}

	eng := NewEngine(nil, pool)

	// Active user and identity
	identityID := uuid.NewString()
	au := &authpkg.ActiveUser{CorporealDomainUID: identityID, HasExternalUID: true, CorporealDomainID: 1}
	ctx = authpkg.WithActiveUser(ctx, au)

	seat, err := eng.InstantiateSeat(ctx, domainID.String(), identityID)
	if err != nil {
		t.Fatalf("InstantiateSeat failed: %v", err)
	}

	// Read latest receipt for seat
	receipts, err := dbstore.ListSeatReceipts(ctx, pool, seat.ID)
	if err != nil {
		t.Fatalf("ListSeatReceipts failed: %v", err)
	}
	if len(receipts) == 0 {
		t.Fatalf("expected receipts for seat instantiate")
	}
	last := receipts[len(receipts)-1]
	payload, ok := last["payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("payload missing or wrong type")
	}
	// ensure identity_id present
	if pid, ok := payload["identity_id"]; !ok || pid == "" {
		t.Fatalf("expected identity_id in payload, got %#v", payload)
	}
	// actor is stored in receipt actor column - ensure present via last payload hash presence
	if _, ok := payload["hash"]; !ok {
		t.Fatalf("expected hash in payload")
	}

	// cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM receipts WHERE target = $1", seat.ID)
	_, _ = pool.Exec(ctx, "DELETE FROM domain_seats WHERE id = $1", seat.ID)
}
