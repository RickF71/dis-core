# 📜 Manifest of the Ledger  
### DIS-CORE v0.8.4 — *The Great Transition*

---

## I. Purpose

The **Ledger** is the living memory of DIS-CORE.  
It replaces static files and external trust with a verifiable, self-contained record of actions, freezes, and authority events.

Where YAML once described structure, the ledger now **remembers** it.  
Every recorded receipt is a point of consensus — a trace of sovereignty etched into code.

---

## II. Principles

### 1. Immutability through Traceability  
Nothing is truly immutable — but every mutation is traced, signed, and witnessed.  
The ledger does not forbid change; it demands proof.

### 2. Sovereignty through Persistence  
Memory is the foundation of autonomy.  
A domain that remembers its actions no longer depends on external validation.

### 3. Minimalism as Strength  
The ledger is not a blockchain.  
It is a single, verifiable sequence of receipts — efficient, local, and human-readable.

### 4. Distributed Trust  
Every domain maintains its own ledger.  
Consensus is emergent, not imposed. Synchronization occurs through export and verification, never through central authority.

---

## III. Implementation Summary

- Backed by a local **SQLite (modernc)** database: `data/dis_core.db`
- Core table: `receipts`
  ```sql
  CREATE TABLE receipts (
      id TEXT PRIMARY KEY,
      actor TEXT,
      action TEXT,
      timestamp TEXT,
      hash TEXT,
      signature TEXT,
      frozen_core_hash TEXT,
      metadata TEXT
  );
