package receipts_test

import (
    "context"
    "encoding/json"
    "testing"

    testdb "dis-core/internal/testdb"
    "dis-core/internal/ledger"

    "github.com/google/uuid"
)

// TestAuthorityReceiptLinkage verifies that a ci.call.v1 receipt can be
// created and an authority_receipts entry may be linked to it. This is a
// focused integration test that exercises the DB schema and insertion paths.
func TestAuthorityReceiptLinkage(t *testing.T) {
    pool := testdb.SetupTestDB(t)
    testdb.MustHaveDB(t, pool)

    ctx := context.Background()

    // Ensure canonical spine present for domain lookups
    if err := testdb.SeedCanonicalSpine(pool); err != nil {
        t.Fatalf("seed spine: %v", err)
    }

    // Find a domain to act as origin (terra)
    var domainID string
    if err := pool.QueryRow(ctx, `SELECT id::text FROM domains WHERE name = 'terra' OR name = 'domain.terra' LIMIT 1`).Scan(&domainID); err != nil {
        t.Fatalf("select terra domain: %v", err)
    }

    // Insert a canonical Phase-9C receipt row (schema used by receipts verification)
    receiptID := uuid.NewString()
    metadata := map[string]any{"test": "authority-link"}
    metaJSON, _ := json.Marshal(metadata)
    if _, err := pool.Exec(ctx, `
        INSERT INTO receipts_9c (id, receipt_type, event_id, policy_ref, issued_by, issued_at, verified, metadata)
        VALUES ($1::uuid, $2, $3, $4, $5, NOW(), true, $6::jsonb)
    `, receiptID, "ci.call.v1", domainID, "policy.ns#v1", "test.actor", string(metaJSON)); err != nil {
        t.Fatalf("insert receipt: %v", err)
    }

    // Insert an authority_receipts entry linking to the receipt
    authID := "rcpt-" + uuid.NewString()
    payloadObj := map[string]any{"linked_receipt": receiptID, "policy_ref": "policy.ns#v1"}
    payloadJSON, _ := json.Marshal(payloadObj)

    hash := ledger.HashString(string(payloadJSON))

    if _, err := pool.Exec(ctx, `
        INSERT INTO authority_receipts (id, domain_id, action, prev_id, payload, hash, policy_digest, created_at)
        VALUES ($1, $2::uuid, $3, NULL, $4::jsonb, $5, NULL, NOW())
    `, authID, domainID, "domain.test.v1", string(payloadJSON), hash); err != nil {
        t.Fatalf("insert authority_receipt: %v", err)
    }

    // Verify the authority_receipts row exists and links back to the receipt
    var found string
    if err := pool.QueryRow(ctx, `SELECT id FROM authority_receipts WHERE payload->>'linked_receipt' = $1 LIMIT 1`, receiptID).Scan(&found); err != nil {
        t.Fatalf("verify authority_receipt link: %v", err)
    }
    if found != authID {
        t.Fatalf("authority_receipt id mismatch: expected %s got %s", authID, found)
    }

    // Sanity-check the original receipts row still exists
    var typ string
    if err := pool.QueryRow(ctx, `SELECT receipt_type FROM receipts_9c WHERE id = $1::uuid LIMIT 1`, receiptID).Scan(&typ); err != nil {
        t.Fatalf("verify receipts row: %v", err)
    }
    if typ != "ci.call.v1" {
        t.Fatalf("unexpected receipt type: %s", typ)
    }

    // cleanup: leaving data to truncated by test harness
}
