-- Create a simple bootstrap_state table to persist bootstrap actor binding
CREATE TABLE IF NOT EXISTS bootstrap_state (
    id SERIAL PRIMARY KEY,
    actor_id TEXT,
    nonce TEXT,
    updated_at TIMESTAMPTZ DEFAULT now()
);
-- Insert initial empty row if none exists
INSERT INTO bootstrap_state (actor_id, nonce)
SELECT '', ''
WHERE NOT EXISTS (SELECT 1 FROM bootstrap_state);
