# COPILOT ONBOARDING — DIS/Finagler Project

This document teaches GitHub Copilot how the DIS architecture works, how Finagler connects to it, and what invariants must always remain true across the entire codebase. Copilot should use this file as a reference when generating backend (Go) or frontend (React) code.

============================================================
SECTION 1 — PROJECT OVERVIEW
============================================================

DIS-Core (Go backend):
- A domain-based sovereignty engine.
- Stores domains, seats, actors, receipts, policies.
- Acts as the authoritative source of truth.
- Communicates via JSON REST APIs.
- Uses Postgres as database.

Finagler (React frontend):
- Visual client for DIS-Core.
- Has two modes:
  (1) FinaglerNone: user is NOT authenticated
  (2) FinaglerProper: user IS authenticated and inside DIS
- Finagler NEVER performs authentication; it only reacts to DIS-Core.

Actor model:
- A PERSON exists only in their corporeal domain.
- All DIS-visible identity happens through ACTORS.
- Each user has a default actor created automatically.
- SuperBar “acting-as” is allowed for testing, but not for production.

============================================================
SECTION 2 — DOMAIN MODEL
============================================================

Domains are hierarchical:
domain.null → child → child → ...

Rules:
- Every domain inherits structure from ancestors.
- A user’s corporeal domain is created from their identity.
- Children may override inherited rules, but cannot erase the ancestor tree.

Required invariants:
- A domain stands by itself.
- To connect into DIS, a domain must connect through a Jikka (websocket).
- Descendant domains “see upward” in schema, but must adopt schemas intentionally.

Post-auth landing:
- After login, user must always be placed inside their CORPOREAL DOMAIN.
- ActiveDomain must be set to the corporeal domain ID.
- ActiveActor must be set to the default actor ID.

============================================================
SECTION 3 — AUTHENTICATION FLOW
============================================================

1. FinaglerNone fetches challenge from DIS-Core.
2. User completes the challenge externally (QR, script, etc.).
3. DIS-Core marks challenge as solved.
4. FinaglerNone polls until authenticated.
5. After authentication:
   - GET /api/me returns user details.
   - Finagler must set:
        activeDomain = me.corporeal_domain_id
        activeActor  = me.default_actor_id
   - Navigate to /domain/{corporeal_domain_id}

Copilot must ALWAYS assume POST-AUTH INITIALIZATION must run.

============================================================
SECTION 4 — DOMAIN CSS SYSTEM
============================================================

Each domain stores its OWN CSS snippet (“domain CSS”):
- This is NOT final CSS.
- It is combined with ancestor CSS to produce “resolved CSS”.

Resolved CSS = concatenation of ancestor CSS from root → leaf.

Frontend:
- Finagler calls: GET /api/domain/{id}/css/resolved
- Injects CSS into <style id="domain-style">
- Switching domain REPLACES domain-style content.

Copilot must never mix “domain CSS” with “resolved CSS.”

============================================================
SECTION 5 — KEY BACKEND ENDPOINTS
============================================================

/api/me
/api/me/actors
/api/me/active-actor (GET/POST)
/api/domain/{id}/css
/api/domain/{id}/css/resolved
/api/domain/{id}/layout
/api/status
/api/identity/list

All endpoints return JSON.
Copilot must preserve these routes exactly.

============================================================
SECTION 6 — FRONTEND CONTEXT LAYERS
============================================================

useActiveUser:
- Stores authenticated user object.

useDomain:
- Stores activeDomain and setActiveDomain.

useActor:
- Stores activeActor and setActiveActor.

useResolvedCSS:
- Automatically fetch CSS for activeDomain.

FinaglerProper is valid ONLY when:
- user != null
- activeDomain != "none"

============================================================
SECTION 7 — CODING GUIDELINES FOR COPILOT
============================================================

BE CONSISTENT WITH:
- Existing filenames
- Existing routing style
- Existing DB patterns
- Existing gin router style in Go

NEVER invent:
- new auth endpoints
- new core DIS concepts
- new domain states
- new actor types
- new schema inheritance behaviors

ALWAYS:
- Use the corporeal domain as the user’s entry point.
- Apply correct domain CSS after domain switch.
- Follow the REST API exactly as defined.
- Keep Go handlers small and composable.
- Match existing SQL schema and naming patterns.

============================================================
END OF FILE — Copilot should ingest this automatically.
============================================================
