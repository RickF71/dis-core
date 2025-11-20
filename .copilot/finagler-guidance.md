# Finagler Guidance — Domain-Oriented UI

Finagler's UI must reflect the new sovereignty model.

Copilot should generate:

- Domain-oriented queries (not universe-oriented)
- Use of /api/domain/<id> to retrieve domain loader output
- Actor switching through domain logic
- Domain navigation based on ancestry tree
- UI components that rely on domain inheritance
- Strict avoidance of references to global terra/numen/lima

Finagler is NOT allowed to assume:

- The existence of system-level domains
- That any domain exists except what DIS-Core returns
