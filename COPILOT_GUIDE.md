# 🤖 DIS-CORE Copilot Guide
**Purpose:** Assist with the Mother Of All Refactors (MOAR Phase 2).
This guide defines how GitHub Copilot should be used across the repository.

---

## 🧭 0. General Principles
1. **Copilot is a power tool, not an author.**
   Human intent (Rick F71) defines structure, Copilot fills in boilerplate.
2. **All autogen code must be human-verified.**
   Include the tag below in every Copilot-assisted file:
   ```go
   // AUTOGEN-COPILOT: initial scaffold, verified by RickF71
   // Ref: MOAR Phase 2 – <component name>
