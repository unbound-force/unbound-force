---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:29:44Z
identity: issue-triage-20261102T002944-jay-flowers
tier: draft
---

Phased-migration triage: a child issue that opens with an explicit "Umbrella: #N" reference is a deliberate per-phase/per-repo decomposition of the umbrella, not a duplicate — especially when the umbrella's own phase-tracking checklist enumerates the child (unbound-force umbrella #797 Phase 3 lists #802/#803/#804/#805 as parallel sibling verification passes). Treat the referenced umbrella as provenance, not as a competing issue. Validity then rests on the child carrying its own concrete goal, bounded scope with an explicit out-of-scope deferral (e.g., #805 defers install/getting-started doc changes to the Phase 5 cutover issue), and a testable done-when clause ("Website repo tooling works under both v1 and v2 binaries with the same committed tree"). Operational nuance: when v1 and v2 binaries share a command name and cannot be installed side by side (a verified fact recorded on #797), a "works under both binaries" done-when clause implies sequential verification runs against the same committed tree, not concurrent installs.
