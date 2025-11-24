package api_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    testdb "dis-core/internal/testdb"
    "dis-core/internal/api"
    "dis-core/internal/ledger"

    "github.com/google/uuid"
)

func TestVerifyReceiptHandler_IncludesVerification(t *testing.T) {
    pool := testdb.SetupTestDB(t)
    testdb.MustHaveDB(t, pool)

    ctx := context.Background()
    if err := testdb.SeedCanonicalSpine(pool); err != nil {
        t.Fatalf("seed spine: %v", err)
    }

    // Insert a Phase-9C receipts row
    receiptID := uuid.NewString()
    if _, err := pool.Exec(ctx, `INSERT INTO receipts_9c (id, receipt_type, event_id, policy_ref, issued_by, issued_at, verified, metadata) VALUES ($1::uuid,$2,$3,$4,$5,NOW(),true,$6::jsonb)`, receiptID, "ci.call.v1", "terra", "policy.ns#v1", "test.actor", `{"test":"x"}`); err != nil {
        t.Fatalf("insert receipt: %v", err)
    }

    // Also insert authority_receipts linking to this receipt
    authID := "rcpt-" + uuid.NewString()
    payload := `{"linked_receipt":"` + receiptID + `","policy_ref":"policy.ns#v1"}`
    hash := ledger.HashString(payload)
    if _, err := pool.Exec(ctx, `INSERT INTO authority_receipts (id, domain_id, action, prev_id, payload, hash, policy_digest, created_at) VALUES ($1, (SELECT id FROM domains WHERE name='terra' LIMIT 1), $2, NULL, $3::jsonb, $4, NULL, NOW())`, authID, "domain.test.v1", payload, hash); err != nil {
        t.Fatalf("insert authority_receipt: %v", err)
    }

    // Start server and call handler via router
    // Create ledger instance for NewWithPolicy
    led, err := ledger.Open(ctx, "", pool, nil)
    if err != nil {
        t.Fatalf("open ledger: %v", err)
    }
    defer led.Close()

    s := api.NewWithPolicy(pool, led, nil, nil)

    req := httptest.NewRequest(http.MethodGet, "/api/receipts/verify/"+receiptID, nil)
    rr := httptest.NewRecorder()
    s.Handler().ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
    }

    var resp map[string]any
    if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
        t.Fatalf("decode response: %v", err)
    }

    if rid, ok := resp["receipt_id"].(string); !ok || rid == "" {
        t.Fatalf("missing receipt_id in response")
    }
    if pr, ok := resp["policy_ref"].(string); !ok || pr != "policy.ns#v1" {
        t.Fatalf("unexpected policy_ref: %v", resp["policy_ref"])
    }
}
