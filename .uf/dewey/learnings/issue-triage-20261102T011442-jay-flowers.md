---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:42Z
identity: issue-triage-20261102T011442-jay-flowers
tier: draft
---

Documentation-gate record issues across distinct changes share heavy keyword overlap ("docs:" prefix, convention-pack rule ids like DE-00X, "pack/skill additions"), so Phase 1 keyword duplicate candidates are frequently false positives. Verified triaging unbound-force/unbound-force#781 (seam-verification, DE-007, parent change #765): the nearest duplicate candidate #791 tracks a different change (delegation-skill-enforcement, DE-008), and the change's own metrics.md duplicate check pre-filing found #762/#720 as nearest neighbors with different scope. Recipe: disambiguate docs-gate records on the parent change reference (#NNN / opsx branch) and the specific pack rule id before classifying duplicate; a shared DE-rule vocabulary alone is a topical match, not duplication.
