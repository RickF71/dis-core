# MOAR X – Reorientation Edition
# Copilot Structural Context
# Purpose: Guide Copilot to generate code aligned with the new DIS-Core sovereign domain model.

## Summary
We are performing a major architectural shift in DIS-Core and Finagler.
The previous model relied on global bootstrap logic that created synthetic
domains, identity triads, prime seats, and top-down governance. This model
is now deprecated.

The new architecture centers on:
- A Universal Domain Loader
- On-demand ancestry resolution
- Per-domain sovereignty rooted at the PSeat
- Minimal global bootstrap (only domain.null)
- No global terra/numen/lima triads
- No synthetic or auto-generated domains
- No global prime-seat sweeps
- No top-down identity or policy injection
- Domains self-govern based on inherited capabilities

Copilot must generate code that supports this architecture.

## What Copilot SHOULD DO
- Support the creation of a Universal Domain Loader.
- Implement recursive ancestry resolution:
  Load domain → ensure PSeat → load parent → merge capabilities.
- Treat domain.null as the only globally bootstrapped domain.
- Use a minimal `main.go` that only initializes:
  - logging
  - config
  - DB
  - schema tables
  - root domain
  - OPA engine
  - authority console
  - StartDaemon()
- Ensure all governance logic (pseats, schema inheritance,
  identity binding, freeze inheritance) happens inside domain loading.
- Encourage domain-centric logic and eliminate server-level governance.
- Generate domain-aware handlers and middleware.
- Assume DIS-Core will evolve toward fully domain-based capabilities.
- Ensure new endpoints and logic respect ancestral capability inheritance.
- Reflect the principle: “A domain can do anything it is capable of within
  its own subtree, according to its PSeat and inherited capabilities.”

## What Copilot MUST NOT DO
- MUST NOT propose terra/numen/lima triad creation or bootstrap.
- MUST NOT propose synthetic or system domains at startup.
- MUST NOT generate GOV-1, GOV-6, or PhaseS0 seat bootstrap code.
- MUST NOT create testmech or auto.corporeal domains globally.
- MUST NOT recreate or reference the old bootstrap identity scaffolding.
- MUST NOT assume that DIS-Core creates global domains for users.
- MUST NOT inject governance at the server level.
- MUST NOT propose global prime-seat sweep logic.
- MUST NOT propose direct DB mutations that bypass receipts or policy.
- MUST NOT suggest modifying domain ancestry or ID structures.

## Required Principles
- Sovereignty flows down the domain tree.
- PSeat = sovereign authority for that branch and descendants.
- Domains inherit capabilities and policy from parents.
- Capabilities merge but ancestor denials take precedence.
- Domain loader = source of structure, not bootstrap.
- DIS-Core evaluates governance; it does not impose governance.

## Key Architectural Notes
- Domain Loader is responsible for:
  - loading domain from DB
  - resolving parent domain(s)
  - ensuring PSeat existence
  - merging inherited capabilities (schemas, gates, policy)
  - resolving freeze state
  - resolving identity bindings inside the tree
- StartDaemon runs only the API server; no governance logic.
- New code should be designed with recursive, domain-first thinking.

## Guidance for Finagler Support
- UI should not assume terra/numen/lima exist.
- UI should rely on the Domain Loader outputs returned from /api/domain.
- Actor switching and identity flows should be domain-centric.
- Finagler is domain-aware; it is not universe-aware.
- Do not generate UI that assumes global domains.

## Philosophy
- DIS-Core is not a central authority.
- DIS-Core evaluates domain-defined authority.
- Domains are sovereign over their subtrees.
- The system is decentralized by construction.
- Bootstrapping governs nothing except domain.null.
