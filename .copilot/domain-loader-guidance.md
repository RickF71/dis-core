# Universal Domain Loader — Guidance for Copilot

All future DIS-Core governance logic must be implemented inside the Universal
Domain Loader. This loader is recursive and sovereign.

Copilot should generate loader code that:

1. Loads a domain from DB
2. Ensures PSeat existence
3. Ensures MSeats/actors as needed
4. Recursively loads parent
5. Merges inherited capabilities
6. Merges adopted schemas
7. Loads applicable OPA policy files
8. Merges freeze state
9. Resolves identity lineage
10. Returns a fully-governable Domain struct

Additional Notes:
- Loader must prevent altering ancestor domain policy
- Loader must guarantee receipt references to ancestry
- Loader must operate on ANY domain (corporeal, corporate, system)
- Loader must not rely on synthetic global scaffolding
