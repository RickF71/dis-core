CREATE TABLE IF NOT EXISTS terra_giveroot_events (
  id TEXT PRIMARY KEY,
  giver_id TEXT NOT NULL,
  applicant_id TEXT NOT NULL,
  method_kind TEXT NOT NULL,
  method_note TEXT,
  ts TIMESTAMP NOT NULL,
  giver_sig TEXT NOT NULL,
  applicant_sig TEXT NOT NULL,
  receipt_hash TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_giveroot_receipt ON terra_giveroot_events(receipt_hash);
