# DIS-CORE v0.8.5 → v0.9 Transition Roadmap

## ✅ Current Checkpoint
**Version:** v0.8.5  
**Status:** Frozen  
**Summary:**  
- Ledger integration (SQLite via `modernc.org/sqlite`) operational  
- Schema registry + freeze hashing validated  
- Domains: `domain.notech`, `domain.government`, `domain.usa` verified  

---

## 🔧 Immediate Next Steps (v0.8.6)
1. Refactor ledger into an interface:
   - `ledger.Store` with `Save()`, `Query()`, `Audit()`, `Purge()`
   - Unit tests for persistence and verification across restarts.
2. Add CLI tools:
   - `disctl ledger list`
   - `disctl ledger verify`
3. Create test data: sample receipts and schema registrations.

---

## 🧩 Schema & Domain Expansion (v0.8.7)
- Auto-register schemas from the ledger on load.
- Implement `ledger.RegisterSchema()` and `ledger.ListSchemas()`.
- Deprecate static `/schemas` as mandatory.
- Add human-readable schema registry dump.

---

## 🏛️ Governance Layer (v0.8.8)
- Add core domain actions:
  - `vote`, `allocate_assets`, `issue_invite`, `accept_invite`
- Generate receipts for each action.
- Extend NOTECH schema to monitor these events.

---

## 🌐 v0.9 — Ledger Sovereignty Milestone
- All schemas and domains registered via ledger.
- Web console for inspecting:
  - Receipts
  - Schemas
  - Domain states
- REST endpoints:
  - `/api/ledger`, `/api/schema`, `/api/domain`

---

## 🪙 Optional Branch: Simula Terra Prototype
After ledger sovereignty stabilizes:
- Start `simula-trade` branch.
- Prototype trade simulator using DIS receipts.
- Player domains: `usa`, `europa`, `limen.commons`, `notech`.

---

### 🧭 Long-Term Goal
DIS becomes self-contained: no static files required to define or validate domains.
