# Jikka Relational Model — A/B/C Structure

Overview
--------
This design note defines the relational model used by Jikka for representing pairwise actor meaning and a shared coherence field. It uses a simple A/B/C set model to express private, shared, and emergent meaning and to guide how JikkaDomain will integrate with DIS mechanics (policy, receipts, freeze) without becoming part of the sovereign domain spine.

Core definitions
----------------
- A — Actor A's meaningful state (private or public values that A treats as meaningful).
- B — Actor B's meaningful state.
- C — Jikka coherence field (the explicit, visible, shared field that Jikka manages).

Fundamental identity
--------------------
- Actual Shared Jikka = A ∪ B ∪ C
  - This expresses the totality of meaning that participates in the shared relational model, whether owned by an actor or by the explicit coherence field C.

Seven-region Venn interpretation
--------------------------------
The full partition of the A/B/C universe yields seven distinct semantic regions. Each region is useful when reasoning about provenance, receipts, and coherence computation.

1. A only
   - Elements meaningful only to A; invisible to B and not represented in C.

2. B only
   - Elements meaningful only to B; invisible to A and not represented in C.

3. C only
   - Elements present in the coherence field C but not attributable to either actor A or B.
   - Interpreted as pure emergent coherence (see below).

4. A ∩ B
   - Meaning shared directly between A and B but not captured in C.
   - Represents private alignment between actors that Jikka has not (yet) incorporated.

5. A ∩ C
   - Meaning A contributed to the visible coherence field.

6. B ∩ C
   - Meaning B contributed to the visible coherence field.

7. A ∩ B ∩ C (mutual coherence)
   - Elements observed and adopted by both actors and present in the shared field — the strongest form of coherence.

Mirror-shadow and emergent regions
----------------------------------
- (A ∪ B) − C — "hidden relational meaning unknown to Jikka"
  - Also called the mirror-shadow region.
  - These are private or unaligned meanings between actors that influence interactions but are not visible to C. They can affect behavior and should be modeled as latent state influencing coherence calculations, even though C cannot directly observe them.

- C − (A ∪ B) — "pure emergent coherence"
  - Meaning that arises in the shared field that cannot be traced to any single actor's private state.
  - This represents emergent patterns or system-level constraints produced by interaction.

Functional roles and high-level semantics
----------------------------------------
- C (visible shared coherence):
  - The canonical, observable coherence field Jikka exposes.
  - Directly used for read operations, shared receipts, and policy-level reasoning about mutual states.

- (A ∪ B) − C (hidden/submerged relational coherence):
  - Private or misaligned relational meaning. Not directly visible to C, but important for modeling intent, friction, and incentives.
  - Treated as latent vectors that influence coherence density (μ) and reconciliation strategies.

- C − (A ∪ B) (potential/emergent coherence):
  - Emergent shared patterns not attributable to either actor alone.
  - Modeled as an independent contributor to coherence density and as candidate content for canonicalization into actor-visible receipts.

Notes for implementation and naming
---------------------------------
- JikkaDomain placement:
  - Jikka should be a domain-adjacent construct named `JikkaDomain`.
  - It must remain separate from the sovereign domain spine (null → terra → numan → lima → corporeal); JikkaDomain is not a sovereign domain itself.

- Domain mechanics usage:
  - `JikkaDomain` can leverage domain mechanics such as policy evaluation, receipts, and freeze semantics for coordination and governance.
  - However, policy and receipt semantics should be scoped carefully to ensure Jikka's constructs do not migrate into the primary domain hierarchy.

- Modeling hidden and emergent regions:
  - Represent the three core components as distinct state vectors:
    - A_vector, B_vector: actor-local state representations (sparse or dense feature vectors)
    - Hidden_vector = (A ∪ B) − C: latent/submerged relational features
    - Emergent_vector = C − (A ∪ B): emergent coherence features
  - Coherence density μ should be a scalar (or small set of scalars) derived from vector combinations and used to drive reconciliation, prioritization, and canonicalization logic.

- Persistence and receipts:
  - Mutual regions (A ∩ B ∩ C, A ∩ C, B ∩ C) should be candidates for mutual receipts referencing both actor anchors and the `JikkaDomain` anchor.
  - Hidden regions should generate private receipts (or encrypted metadata) that influence reconciliation but are not published into C without canonicalization.

Integration with TAG & DIS concepts
----------------------------------
- TAG (constructive vs destructive interference):
  - Constructive interference: overlapping vectors across A, B, and C increase μ (coherence density) and produce positive reinforcement toward canonical shared content.
  - Destructive interference: conflicting private vectors (A only vs B only or divergent A ∩ B) reduce μ and increase reconciliation effort.

- DIS (mutual vs private receipts):
  - Mutual receipts: represent registry entries for A ∩ B ∩ C and A ∩ C / B ∩ C where provenance is clear and shared.
  - Private receipts: represent (A ∪ B) − C and A only / B only regions; these inform policy decisions and reconciliation but are not exposed as shared canonical proofs until promoted.

Behavioral expectations and next steps
------------------------------------
- Coherence computation should treat Jikka as a coordination layer that computes μ from actor state vectors, hidden_vector, and emergent_vector.
- Reconciliation flows:
  - Detect regions with low μ where destructive interference is present -> surface recommendations or policy triggers.
  - Promote emergent coherence (C − (A ∪ B)) to actor-visible receipts only after policy checks and mutual agreement patterns are detected.

Concise summary
---------------
- Model: Actual Shared Jikka = A ∪ B ∪ C
- Seven regions map the landscape of private, shared, and emergent meaning.
- Treat hidden and emergent regions as separate state vectors that influence coherence density μ.
- `JikkaDomain` is domain-adjacent: use domain mechanics but remain outside the sovereign domain spine.

This document prioritizes naming, structure, and intended behavior; it is ready for translation into a `JikkaDomain` implementation and concrete types for vectors, receipts, and coherence metrics.
