---
tag: change-management
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:57:02Z
identity: change-management-20261107T195702-jay-flowers
tier: draft
---

Review-phase fixes that add a new executable or user-visible surface must be folded into the change's internal record (a tasks.md entry, a design risk-register or decision sentence, and the documentation-issue scope) in the same commit, matching the fold-back treatment of their sibling fixes; otherwise the archived change record under-describes what shipped even when CHANGELOG and user docs are current. Precedent: PR #957 round 3 — the daily scheduled transitive scan job landed in ci_scheduled.yml with CHANGELOG and docs/osv-transitive-scan.md coverage, but tasks.md 4.4 still named only ci_security.yml, the design risk register still described the compensating scan as push/PR-only, and the #964/#382 documentation issues were not extended to the grown scope, while sibling fixes in the same commit (tasks 4.9, 4.10) were folded back.
