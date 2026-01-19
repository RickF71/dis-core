# dis-core-rust v0.9.0 — Sprint TODO

> **Sprint theme:** Stop raking the whole yard. One strip, end to end.
>
> **Invariant:** If a task does not move one leaf from *exists* → *proven*, it is out of scope.

---

## Sprint Scope (Hard Boundary)

**IN SCOPE**
- dis-core (Rust)
- totem IPC (minimal, stable)
- One end-to-end intent → receipt → persistence flow
- Refusal paths (freeze / deny / break-glass)

**OUT OF SCOPE**
- UI polish (Finagler visuals)
- New domain metaphors
- Performance tuning
- Multiple transports
- Abstractions not forced by repetition

---

## Phase 1 — Ground Flatness (Boot & Stability)

☐ dis-core boots deterministically (no flags, no env tricks)
☐ totem boots deterministically
☐ Startup order defined and documented
☐ Single IPC path established (dis-core → totem or totem → dis-core, not both)
☐ No background tasks without explicit lifecycle ownership

**Exit condition:**
> "It starts, it stays up, it talks."

---

## Phase 2 — Structural Reality Only (No Behavior)

☐ Domain structs exist with zero embedded behavior
☐ Artifact structs exist as pure records
☐ Canonical ID types compile and round-trip
☐ Schemas load and reject invalid input
☐ Identity binding is visible in data (even if naive)

**Forbidden:**
- Implicit meaning
- Auto-creation
- Side effects on load

**Exit condition:**
> "We can store reality without interpreting it."

---

## Phase 3 — One End-to-End Strip (Golden Path)

☐ Select **one** intent type (name it explicitly)
☐ Intent accepted via one entry point
☐ Intent produces exactly one receipt (ci.call.v1 or equivalent)
☐ Receipt is persisted
☐ Receipt can be read back and verified

**Constraints:**
- No branching
- No alternatives
- No retries

**Exit condition:**
> "This one thing happened, and DIS can prove it."

---

## Phase 4 — Refusal Before Power

☐ Deny path exists for the chosen intent
☐ Domain freeze prevents the intent
☐ Break-glass override bypasses freeze **with logging**
☐ Refusal reasons are structured (not strings)
☐ Refusal produces receipts

**Rule:**
> If "no" is weaker than "yes", stop.

**Exit condition:**
> "The system can say no safely."

---

## Phase 5 — Minimal Human Visibility (Not Polish)

☐ One read-only API or log view shows:
  - intent
  - receipt
  - decision
☐ Errors are intelligible to a human
☐ No new authority logic introduced

**Explicitly not UI polish.**

**Exit condition:**
> "A human can understand what already exists."

---

## Phase 6 — Repeat the Same Rake (Optional Stretch)

☐ Add a **second** intent using the exact same path
☐ No new abstractions unless duplication forces it
☐ Confirm refusal logic applies unchanged

**Exit condition:**
> "We are confident this generalizes."

---

## Optional Work Chunks (Non-Timeboxed)

> These are **ordering aids**, not deadlines.
> Take one chunk in an hour, a day, or a week. Skip around **only** if exit conditions remain true.

### Chunk A — Boot Reality
- Covers: Phase 1
- Question it answers: *“Can this thing exist calmly?”*

### Chunk B — Structural Truth
- Covers: Phase 2
- Question it answers: *“What exists, without opinion?”*

### Chunk C — The Golden Strip
- Covers: Phase 3
- Question it answers: *“Can DIS prove a single thing happened?”*

### Chunk D — Saying No
- Covers: Phase 4
- Question it answers: *“Can DIS refuse power safely?”*

### Chunk E — Human Witness
- Covers: Phase 5
- Question it answers: *“Can a human see and trust what happened?”*

### Chunk F — Confidence Check (Optional)
- Covers: Phase 6
- Question it answers: *“Will this survive repetition?”*

---

## Sprint Completion Definition

The sprint is **done** when:
- One intent is provable
- One refusal is provable
- One override is provable
- dis-core feels boring

> Boring means stable. Stable means sovereign.

---

## Parking Lot (Write Down, Do Not Touch)

- Performance
- UI elegance
- Domain creativity
- Multi-transport
- Economic models
- Philosophy expansions

Write it here. Do not act on it during this sprint.
