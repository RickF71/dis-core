# MOAR X — Reorientation Edition
# Architectural Overview for Copilot

DIS-Core is transitioning from a server-centric bootstrap model to a
domain-centric sovereignty model. The Universal Domain Loader becomes
the central engine. Bootstrap should only create domain.null.

Copilot must ensure:

- No synthetic domains are created at startup
- No Terra/Numen/Lima identity triads are injected
- No GOV-1, GOV-6, or PhaseS0 logic returns
- Domain loader handles:
  - ancestry resolution
  - PSeat instantiation
  - schema inheritance
  - policy inheritance
  - freeze inheritance
  - identity lineage
  - capability propagation

Summaries:
- Bootstrap = minimal
- Domain = source of governance
- Loader = evaluates structure
- DIS-Core = evaluates domain authority, not imposes it
