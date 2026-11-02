---
tag: documentation-gate
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:57:11Z
identity: documentation-gate-20261101T215711-jay-flowers
tier: draft
---

Documentation-gate applicability for defect fixes in unbound-force/unbound-force turns on observable workflow deltas, not on where the changed code lives. The accept-variantless-parent-model fix lives in plugin internals (.opencode/plugins/review-dispatch plus its internal/scaffold mirror), yet it changes observable /uf.review-pr, /uf.review-council, and /uf.unleash Step 6 behavior — finalization now succeeds for variantless parent models instead of failing closed INCONCLUSIVE with "requires resolved parent model and variant" — so the AGENTS.md current-repo documentation-issue gate applies and the internal-refactoring exemption does not. Triage categorization for such issues: enhancement (the remaining actionable work is documentation), not bug (the defect is already fixed; the change artifacts record the fix as shipped, tested, and council-approved). Structural note: the documentation gate is a citation pair — the council fix loop for this change demanded "both documentation-gate issue citations" in the OpenSpec artifacts, i.e. the current-repo docs issue (unbound-force/unbound-force #891) plus the unbound-force/website tracking issue required by the org constitution's Cross-Repo Documentation rule. A Curator triaging a current-repo docs-gate issue should expect the website companion to be cited in the change artifacts even when the issue body itself does not reference it, and should not mark the issue invalid for that omission.
