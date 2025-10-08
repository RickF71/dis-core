# DIS-PERSONAL Changelog

## v0.7 — Authority Console (2025-10-08)
### Summary
A major release completing the Authority Console layer, bridging local sovereignty with peer-to-peer verification.  
Bound to **DIS-CORE v1.0 (frozen schema)**.

### Added
- 🔁 **Automated 30-minute verification cycle** — scheduled background audits that detect new receipts and self-verify integrity.
- 🪶 **Heartbeat publisher** — broadcasts the latest verified receipt to trusted peers for distributed synchronization.
- 🌍 **Peer ingestion endpoint** `/api/verify/external` — accepts and validates incoming verification receipts.
- 📜 **Trust ledger** (`versions/v0.7/ledger/trust.json`) — persistent event log for peer exchanges, tracking network trust and status.
- 🧭 **Network configuration file** (`versions/v0.7/network.yaml`) defining trusted peers and endpoints.

### Changed
- Unified console startup with auto-verification trigger.
- Improved startup logs and deterministic timestamps for verification state loading.
- Enhanced receipt provenance output (consistent `by:` and `verified_at` fields).

### Fixed
- Graceful handling of missing verification receipts and invalid peers.
- Avoided redundant re-verification runs on startup.
- Corrected path resolution for internal ledger and schema consistency.

### Status
Stable. Recommended baseline for distributed network testing.  
Next milestone: **v0.8 — Peer handshake and trust synchronization**.
