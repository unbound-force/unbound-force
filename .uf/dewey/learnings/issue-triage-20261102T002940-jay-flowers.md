---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:29:40Z
identity: issue-triage-20261102T002940-jay-flowers
tier: draft
---

Phase-scoped issues inside a phased migration are self-disambiguating in triage when they open with "Umbrella: #NNN" and "Phase: N — <description>" lines and carry a testable "Done when" criterion. The umbrella reference is verifiable against the umbrella issue's phase checklist (e.g., #797 Phase 3 lists each sibling verification pass by number), which distinguishes such issues from duplicates of the umbrella itself. The done-when clause (e.g., "tooling works under both binaries with the same committed tree") anchors otherwise soft verbs like "verify ... behave" and supports an objective VALID verdict with no clarification round-trip. Triage of such issues reduces to a structural completeness check: umbrella link, bounded scope, explicit out-of-scope deferrals to later phases, and correct label already applied.
