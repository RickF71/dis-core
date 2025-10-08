# DIS-CORE v0.6 — Authority Console & Temporal Integrity
**Release Date:** 2025-10-08  
**Core Hash:** `15b437484377ac63cdb227b4fa264010aec06759f5808c699768cbe112f3c930`

---

### ✨ New Features
- **Authority Console API**
  - `/api/console/action` → executes domain actions and generates signed receipts
  - `/api/receipts` and `/api/receipts/{id}` → browse issued receipts
  - `/api/verify/all` → runs verification sweeper and returns signed `domain.verify.v1` receipt

- **Automated Verification Scheduler**
  - Runs every 30 minutes
  - Verifies all receipts under `versions/v0.6/receipts/generated/`
  - Issues self-signed integrity receipts (`domain.verify.v1`)

- **Smart Skip Logic**
  - Detects when no new receipts have been added or modified
  - Skips redundant verification cycles, reducing overhead

- **Persistent Verification State**
  - `last_verification.txt` stores the timestamp of the last successful audit
  - Reloaded automatically on startup so the node remembers its previous state

- **Filesystem-based Awareness**
  - Verification tied to actual receipt modification times
  - 5-second tolerance buffer prevents near-simultaneous false triggers

---

### 🔐 Security & Integrity
- All receipts signed with Ed25519
- Provenance chain includes SAT, domain, and policy refs
- Verification receipts include embedded public key for independent checking

---

### 🧠 Design Principles
- Deterministic, auditable state changes
- Minimal background overhead (< 1 ms per run)
- No redundant ledger writes during idle periods
- “Do nothing when stable” behavior — first expression of autonomous sovereignty

---

### 📁 Directory Layout
