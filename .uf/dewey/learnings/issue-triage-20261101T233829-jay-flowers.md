---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:38:29Z
identity: issue-triage-20261101T233829-jay-flowers
tier: draft
---

Documentation-gate issues are deterministically triageable when they carry four elements, as unbound-force issue #827 (docs sync for fix-init-speckit-sentinel) demonstrates: (1) the governing gate they satisfy (AGENTS.md documentation gate for a user-facing change), (2) the exact pages needing updates, (3) concrete behavior deltas verifiable against the change's spec artifacts (sentinel path `.specify/scripts`, exact invocation flags, visible vs silent skip, config escape hatch), and (4) explicit relationships to prior issues (supersedes #478 with rationale; #503 declared separate scope). Verification then reduces to three cheap checks: the change directory exists (openspec/changes/&lt;name&gt;/), the named pages exist, and the stale wording the issue quotes is present at the cited locations (architecture.md "silently skipped"; cli-reference.md dewey-only `method: skip` note). A generic mechanism grep (scaffold.go `cfg.Setup.Tools[name].Method == "skip"`) grounds escape-hatch claims. Issues missing element 3 or 4 typically warrant NEEDS-CLARIFICATION; issues carrying all four triage cleanly as VALID/objective/enhancement with no split — one behavior change across multiple doc pages is cohesive single-PR scope, not a split candidate.
