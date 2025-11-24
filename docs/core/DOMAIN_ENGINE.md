# DIS Domain Engine — Architectural Contract (v1)

The Domain Engine is the central organizing component of DIS-Core.
All operations in DIS must occur *within* the context of a Domain.
All receipts originate from the domain where the actor performs the action.
All policy evaluation is scoped to a domain.
All actions produce domain-scoped receipts (envelopes with panels).
No DIS operation should occur without an associated domain_id.

## Domain Definition
A Domain is the fundamental container of:
  - meaning
  - policy
  - identity flows
  - receipts (local chain)
  - dimensional location
  - parent/child lineage

Fields (canonical):
  - ID (string/UUID)
  - Name
  - Dimension (0D..corporeal)
  - ParentID
  - Type
  - Metadata
  - Capabilities (schema-derived)

## Domain Engine Responsibilities
  - Resolve domains from ID or context.
  - Route actions to domain logic.
  - Apply policies in domain context.
  - Emit domain-local receipts (envelopes).
  - Manage dimensional spine transitions.
  - Track domain freeze/unfreeze state.
  - Provide a unified interface for all DIS operations.

## Domain Engine Non-Responsibilities
  - HTTP concerns (API layer only adapts requests).
  - UI concerns.
  - Direct DB access outside of domain-aware wrappers.

## Core Contract
DIS-Core *is* itself a domain engine.
Every piece of business logic must migrate behind the Engine
as the project evolves through MR-4 and beyond.

# End of DOMAIN_ENGINE.md
