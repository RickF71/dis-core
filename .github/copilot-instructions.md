/task Implement Phase 9C: Receipt Verification & Provenance Continuity

Objectives:

1. Create migration file: db/migrations/20251110_add_receipts_table.sql
   SQL:
   CREATE TABLE IF NOT EXISTS receipts (
       id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       receipt_type    TEXT NOT NULL,
       event_id        TEXT NOT NULL,
       policy_ref      TEXT,
       redaction_ref   TEXT,
       issued_by       TEXT,
       issued_at       TIMESTAMPTZ DEFAULT now(),
       verified        BOOLEAN DEFAULT FALSE,
       metadata        JSONB DEFAULT '{}'::jsonb
   );
   CREATE INDEX IF NOT EXISTS idx_receipts_event_id ON receipts(event_id);
   CREATE INDEX IF NOT EXISTS idx_receipts_policy_ref ON receipts(policy_ref);
   CREATE INDEX IF NOT EXISTS idx_receipts_redaction_ref ON receipts(redaction_ref);
   CREATE OR REPLACE VIEW receipts_orphan_view AS
   SELECT id, receipt_type, event_id, issued_at
   FROM receipts
   WHERE policy_ref IS NULL OR redaction_ref IS NULL;
   COMMENT ON TABLE receipts IS 'Stores all DIS receipts for provenance and redaction continuity verification (Phase 9C).';

2. Add new Go model: internal/receipts/model.go
   package receipts

   import "time"

   type Receipt struct {
       ID           string         `json:"id"`
       ReceiptType  string         `json:"receipt_type"`
       EventID      string         `json:"event_id"`
       PolicyRef    string         `json:"policy_ref"`
       RedactionRef string         `json:"redaction_ref"`
       IssuedBy     string         `json:"issued_by"`
       IssuedAt     time.Time      `json:"issued_at"`
       Verified     bool           `json:"verified"`
       Metadata     map[string]any `json:"metadata"`
   }

   type VerificationResult struct {
       ReceiptID     string   `json:"receipt_id"`
       Verified       bool     `json:"verified"`
       PolicyRef      string   `json:"policy_ref"`
       RedactionRef   string   `json:"redaction_ref"`
       Timestamp      string   `json:"timestamp"`
       Issues         []string `json:"issues"`
   }

3. Implement verification logic: internal/receipts/verify.go
   package receipts

   import (
       "context"
       "fmt"
       "time"
       "github.com/jackc/pgx/v5/pgxpool"
   )

   func VerifyReceipt(ctx context.Context, pool *pgxpool.Pool, id string) (VerificationResult, error) {
       var r Receipt
       var result VerificationResult
       err := pool.QueryRow(ctx, `
           SELECT id, receipt_type, event_id, policy_ref, redaction_ref, issued_by, issued_at, verified
           FROM receipts WHERE id = $1
       `, id).Scan(&r.ID, &r.ReceiptType, &r.EventID, &r.PolicyRef, &r.RedactionRef, &r.IssuedBy, &r.IssuedAt, &r.Verified)
       if err != nil {
           return result, fmt.Errorf("receipt not found: %w", err)
       }

       result.ReceiptID = r.ID
       result.PolicyRef = r.PolicyRef
       result.RedactionRef = r.RedactionRef
       result.Timestamp = time.Now().UTC().Format(time.RFC3339)

       if r.PolicyRef == "" {
           result.Issues = append(result.Issues, "missing policy_ref")
       }
       if r.RedactionRef == "" {
           result.Issues = append(result.Issues, "missing redaction_ref")
       }
       result.Verified = len(result.Issues) == 0
       return result, nil
   }

4. Add API handler in internal/api/receipts_verify.go
   func (s *Server) handleVerifyReceipt(w http.ResponseWriter, r *http.Request) {
       ctx := r.Context()
       id := chi.URLParam(r, "id")
       result, err := receipts.VerifyReceipt(ctx, s.db, id)
       if err != nil {
           http.Error(w, err.Error(), http.StatusNotFound)
           return
       }
       json.NewEncoder(w).Encode(result)
   }

5. Register route in internal/api/router.chi.go
   mux.Get("/api/receipts/verify/{id}", s.handleVerifyReceipt)

6. During bootstrap, after authority console initialization:
   - Run continuity check:
     SELECT COUNT(*) FROM receipts WHERE policy_ref IS NULL OR redaction_ref IS NULL;
   - Log output to phases/phase_9c.log:
     "✅ Phase 9C — Verified {count} receipts, {orphans} orphan entries (timestamp)"

7. Extend /api/authority/schema response:
   Add field:
     "receipts": ["ci.call.v1", "ci.import.v1"]

Expected Outcome:
- Receipts table exists with indexes and continuity view.
- /api/receipts/verify/{id} endpoint returns full verification details.
- Provenance continuity checked at bootstrap.
- Authority Console schema includes receipt mappings.
- phase_9c.log created with verification summary.
