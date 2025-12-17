# DIS-Core

**DIS-Core** is the foundational runtime for the **Direct Individual Sovereignty (DIS)** system.

It implements the lowest, non-bypassable authority boundary of DIS: a kernel that evaluates attempted actions (commits) against identity and domain — and **denies by default**.

This repository contains the **first stable spine** of DIS.

---

## What DIS-Core Is

DIS-Core is **not** an application, a UI, or a policy engine.

It is:

- A **typed authority kernel**
- A **single commit choke point** for domain mutations
- A **deny-by-default enforcement layer**
- A **witness of attempted coherence** between identity and domain

All higher-level behavior (policy, allowance, receipts, UI, governance) is built *on top* of this core.

(opinion) If DIS were a legal system, DIS-Core would be the court clerk — not the judge, jury, or legislature.

---

## Current Status

**Version:** v0.1.0  
**State:** Stable foundation, minimal surface

Implemented:

- Typed `DomainId` and `IdentityRef`
- Kernel commit boundary (`Kernel::commit`)
- Deny-by-default decision model
- Capsule traversal plumbing (Layer6 spine)
- HTTP API for commit testing
- Clean separation between kernel and API layers

Not yet implemented (by design):

- Policy engines
- Allow rules
- Persistent receipts
- Economic or incentive logic
- UI concerns

---

## Architecture Overview

Finagler (UI) communicates with DIS-Core over HTTP.

Authority lives **only** in the kernel.

- UI and clients request actions
- API transports requests
- Kernel evaluates and witnesses
- Silence defaults to denial

Key principles:

- **APIs do not mint authority**
- **Identities are explicit and typed**
- **All mutations pass through a single choke point**
- **Ambiguity never grants permission**

---

## Example: Commit Attempt

Test a commit against the kernel:

    curl -X POST http://localhost:8787/api/commit/test

Example response:

    {
      "outcome": "denied",
      "decision_ref": "dec:DomainId(...):Actor { id: \"actor.test\" }:Decision { allow: false, reason: DenyByDefault }"
    }

This response demonstrates:

- The kernel witnessed the attempt
- Identity and domain were evaluated
- No implicit authority was granted
- The system behaved deterministically

(opinion) At this stage, **“Denied” is a successful outcome**.

---

## Project Philosophy

DIS is built on a few non-negotiable ideas:

- Humans remain the ultimate source of authority
- Systems may amplify power, but must not originate it
- Identity, consent, and action must be traceable
- Silence and ambiguity must never grant permission

DIS-Core enforces these rules mechanically.

---

## Repository Scope

This repository intentionally contains **only**:

- The DIS kernel
- Minimal API wiring
- Structural primitives (domains, identities, capsules)

Related components live elsewhere:

- **Finagler** — UI and visualization
- **DIS-Bridge** — integration layer
- **Policy packs** — domain-specific rules
- **Economic models** — future work

---

## Building & Running

### Prerequisites

- Rust (stable)
- Cargo

### Build

    cargo build

### Run

    cargo run

Server starts at:

    http://localhost:8787

---

## Versioning

This repository uses semantic versioning.

- **v0.1.0** — Initial stable kernel and commit boundary
- Future versions will expand capability without breaking authority guarantees

---

## License

TBD — intentionally deferred until governance and policy models stabilize.

---

## Final Note

DIS-Core is intentionally **boring** right now.

That’s a feature.

(opinion) Authority systems should only become interesting *after* they become correct.
