---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:45:49Z
identity: test-quality-20261105T164549-jay-flowers
tier: draft
---

Spec acceptance scenarios that describe cross-mechanism file-survival outcomes need one composed acceptance test in addition to piecewise mechanism guards. In PR #957 (stop-shipping-plugin-lock), the SC-FR-003 scenario "Lock written by another tool" — a foreign .opencode/package-lock.json must survive the next uf init byte-identical — was enforced only piecewise (isToolOwned/isDivisorAsset table flips, embed-absence checks, --no-package-lock command pins); a repo-wide grep confirmed no test writes a foreign lock into a scaffold target before Run(). Piecewise guards miss conjunction regressions (re-embedding the asset plus a predicate flip back to tool-owned) and leave design invariants like "uf never deletes a user-owned lock" (design D3) unguarded — a future stale-asset cleanup pass would keep every piecewise test green while deleting user locks. The composed test is cheap and belongs in the Go scaffold suite (make test, not make plugin-test, per learning/scaffold-parity-20261020T113858-jay-flowers): t.TempDir target, sentinel bytes, Run() via the injected-command seams, assert byte-identity.
