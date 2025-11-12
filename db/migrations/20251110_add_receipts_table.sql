-- Migration for Phase 9C: Receipt Verification & Provenance Continuity
-- File: db/migrations/20251110_add_receipts_table.sql

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
COMMENT ON COLUMN receipts.id IS 'Unique UUID identifier for the receipt';
COMMENT ON COLUMN receipts.receipt_type IS 'Type of receipt (ci.call.v1, ci.import.v1, etc.)';
COMMENT ON COLUMN receipts.event_id IS 'ID of the event this receipt documents';
COMMENT ON COLUMN receipts.policy_ref IS 'Reference to the policy that governed this event';
COMMENT ON COLUMN receipts.redaction_ref IS 'Reference to redaction policy if applicable';
COMMENT ON COLUMN receipts.issued_by IS 'Identity that issued this receipt';
COMMENT ON COLUMN receipts.issued_at IS 'Timestamp when receipt was issued';
COMMENT ON COLUMN receipts.verified IS 'Whether receipt has been verified';
COMMENT ON COLUMN receipts.metadata IS 'Additional metadata in JSONB format';
COMMENT ON VIEW receipts_orphan_view IS 'View showing receipts missing policy or redaction references';
