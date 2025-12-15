# DIS_TIME.md
## DIS Temporal Canon

Status: Canonical
Applies to: DIS Spine, Domains, Seats, Artifacts, Events, UI

---

## 1. Principle of Internal Time

DIS does not use wall-clock time for ordering, authority, or causality.

All ordering, commitment, and irreversibility are governed by a deterministic
internal time system defined by progression through the DIS spine.

External time sources may be attached only as witnesses and never affect order.

---

## 2. DIS Tick (Global Time)

A **DIS Tick** is the single observable unit of time in the DIS universe.

- Represented as an unsigned integer: `dis_tick`
- Begins at `1`
- Increments exactly once per full spine traversal
- Never skips
- Never rewinds
- Never advances partially

DIS Tick is the only time exposed to humans and UIs.

---

## 3. The 6D Spine Clock

Each DIS Tick is resolved through exactly six ordered spine phases.
These phases form the internal **6D clock**.

Phases are numbered starting at `1`.

| Phase | Name        | Meaning                          |
|------:|-------------|----------------------------------|
| 1     | Nullus      | Existence / identity anchoring   |
| 2     | Aether      | Connectivity / returnability     |
| 3     | Terra       | Presence / locality              |
| 4     | Numen       | Meaning / salience               |
| 5     | Lima        | Processing / incorporation       |
| 6     | Corporeal   | Commitment / reality lock        |

Phase `0` does not exist.

---

## 4. Zero Law (Outside DIS)

- `dis_tick = 0` is outside the DIS universe
- `phase = 0` is outside the DIS universe

These values represent non-existence and must never appear in runtime state.

---

## 5. Phase Execution Rules

- Every DIS Tick executes all six phases in order
- Phases may not be skipped or reordered
- Phases are deterministic
- Phases are synchronous
- No domain may introduce its own phase ordering

---

## 6. The Commit Law (Corporeal Boundary)

**Corporeal is the only commit boundary.**

Only during the Corporeal phase may the system:
- Create irreversible artifacts
- Modify persistent storage
- Transfer value
- Realize or revoke corporeal seats
- Emit final receipts or ledgers

All other phases may only stage intent or computation.

---

## 7. Domains and Time

Domains do not own independent clocks.

A domain participates in time only by participating in a DIS Tick.

Domains may:
- Observe the current DIS Tick
- React during their designated phase
- Stage internal state

Domains may not:
- Advance time
- Observe wall-clock time
- Commit outside Corporeal

---

## 8. Seats (Canonical Distinction)

### 8.1 Personal Seat (pseat)

Each person owns exactly one **personal domain** (pseat domain).

Within the pseat domain:
- The pseat holder is seated at **all spine layers**
- From Nullus through Corporeal
- These seats are intrinsic, permanent, and non-revocable

This is what makes the domain “personal”.

Storage, secrets, and artifacts belong to the pseat domain
and are committed at the Corporeal layer.

---

### 8.2 Corporeal Seat Projection

When a pseat holder interacts with a **non-personal domain**:

- The pseat itself does not move
- Only a **Corporeal seat** is projected into the target domain
- This projection exists only at the Corporeal layer
- It is domain-scoped, time-scoped, and revocable by that domain

No identity, storage, or root secrets leave the pseat domain.

---

## 9. Seat Presence

A seat may be **defined** or **undefined**.

- Defined: an actor is present; Corporeal seat is inhabited
- Undefined: no actor present; seat persists but cannot act

Seat presence affects permission, not existence.

---

## 10. Event Time

Every irreversible event is bound to:

- `dis_tick`
- `SpinePhase::Corporeal`

Two events cannot collide in time unless they are the same event.

Optional wall-clock timestamps may be attached as witnesses only.

---

## 11. Invariants (Non-Negotiable)

- There is exactly one DIS Tick
- Nullus = 1
- Corporeal seals reality
- Zero never appears in runtime
- Seats do not advance time
- Domains do not advance time
- Participation, not observation, advances time

---

## 12. UI Truth Requirement

Any UI (including Finagler) must:
- Display DIS Tick as primary time
- Never imply wall-clock authority
- Never show partial or speculative ticks
- Reflect Corporeal as the moment of commitment

---

## End of Canon
