# DIS-Core Ledger

## Overview

The **ledger** is the authoritative record-keeping layer for **DIS-Core**.
It provides durable, queryable persistence for every domain event, policy decision, and consent receipt in the system.

All actions in DIS originate from **domains**, but must carry **human authorization** traced to the domain’s root seat.
The ledger exists to preserve those traces — ensuring that no action can occur without a verifiable chain of consent.

---

## Responsibilities

- Maintain **canonical event records** (`ci.call.v1`, `domain.freeze.v1`, etc.)
- Store **policy decisions** from the OPA engine (`policy_decisions` table)
- Enforce **immutability** of signed receipts and audit trails
- Expose simple Go APIs for reading and writing domain-linked data
- Serve as the provenance backbone for DIS authority, governance, and accountability

---

## Core Tables

| Table | Purpose |
|-------|----------|
| **policy_decisions** | Results of evaluated policies (domain actor, human authorizer, action, reason, metadata) |
| **receipts** | Canonical record of consented actions (`ci.call.v1`) with signatures and redaction provenance |
| **events** | Low-level domain state changes (freeze/unfreeze/override, etc.) |
| **canon_files** | Stored canonical definitions and imports for reconciliation |
| **bootstrap_files** | Initial YAML and schema bootstrap artifacts |

---

## Data Model: Domain + Human Authorization

Every record written to the ledger follows the same dual-actor structure:

| Field | Description |
|-------|--------------|
| `actor_domain` | The domain initiating the action |
| `authorized_by` | The human occupant of that domain’s root seat |
| `target_domain` | The domain being acted upon (optional) |
| `action` | The action performed (`domain.freeze.v1`, etc.) |
| `decision` | Boolean allow/deny (for policy events) |
| `reason` | Textual explanation or denial reason |
| `metadata` | JSONB payload (risk score, Rego module info, provenance) |

This guarantees that **every decision** is both *machine-verifiable* and *human-traceable*.

---

## Integration Points

- **OPA / PolicyEngine** — calls `Ledger.SaveDecision()` after every evaluation
- **Receipts Module** — appends verified `ci.call.v1` entries
- **Freeze / Gates / Risk Policies** — log automatic domain enforcement outcomes
- **Finagler Console** — reads from the ledger to render audit and freeze states

---

## API Functions (Go)

| Function | Description |
|-----------|--------------|
| `Open(dsn string, db *sql.DB, reg *schema.Registry)` | Initialize ledger with a DB connection |
| `SaveDecision(*PolicyDecisionRecord)` | Insert policy evaluation results |
| `LogReceipt(*ReceiptRecord)` | Store signed consent receipts |
| `ListDecisions(limit int)` | Query recent policy decisions |
| `Close()` | Graceful cleanup |

---

## Principles

1. **Immutability:**
   Once a decision or receipt is recorded, it cannot be altered — only superseded.

2. **Traceability:**
   Every action has a visible path from domain → human → event → receipt.

3. **Transparency without Exposure:**
   PII is always redacted at source; only provenance tokens and hashes are stored.

4. **Non-Bypass Rule:**
   No part of DIS may modify external state without a corresponding ledger entry.

---

## Example Flow

1. Domain `domain.usa` requests `domain.freeze.v1`
2. Policy engine evaluates OPA modules → `allow = false`, reason: "risk threshold exceeded"
3. Ledger saves decision:
   ```json
   {
     "actor_domain": "domain.usa",
     "authorized_by": "id-rick",
     "action": "domain.freeze.v1",
     "decision": false,
     "reason": "risk threshold exceeded"
   }

4. Finagler Console displays denial with receipt reference.

    Future Extensions

    Policy Tracebacks: link decisions to the exact OPA query & Rego source hash

    Seat Registry Integration: verify human consent via seat lineage lookup

    Freeze TTLs: automatically expire domain freezes after a set duration

    Receipts to SAT (Source Attribution Tokens): tie decisions to globally verifiable proofs

    Status

    v0.9.2 — Ledger subsystem aligned with MinSet-5 kernel
    Identity binding and human authorization model active.
