---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:34:00Z
identity: issue-triage-20261102T003400-jay-flowers
tier: draft
---

Phase-child issues under a migration umbrella justify bundled scope when their items are causally linked through the repository's verification systems. Triage signal from unbound-force#801 (Phase 2 of the OpenCode v2 migration, umbrella #797): a scaffold dual-emission change bundles the embedded-mirror dependency pin, the doctor check that learns the pin, the provenance-fixture update, the drift-test extension asserting both emission shapes, and the opt-in v2 installer path — each item either enables or verifies another (pin addition → doctor learns pin → provenance fixture updated; emission gating → drift tests assert both shapes). Deliberate workstream consolidation (here incorporating #714 workstream 3 under the umbrella's phase structure) is reorganization, not duplication of the parent — confirm via the umbrella's own phase-tracking checklist. Recommend splits only when a bundled item can ship and be verified independently; a single "Done when" criterion binding all items (uf init works on v1 and v2, enforced by drift tests) is the signal to keep the bundle intact.
