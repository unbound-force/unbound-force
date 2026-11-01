---
tag: golden-tests
author: jay-flowers
category: gotcha
created_at: 2026-10-31T13:31:04Z
identity: golden-tests-20261031T133104-jay-flowers
tier: draft
---

Golden-label refresh tasks must derive their location set from a fresh repo-wide grep, not from a remembered file list of the test files. In unbound-force, uf.*.md:<line> site labels live in three locations — .opencode/test/golden-output-shape.test.ts, .opencode/test/dispatch-review-runs.test.ts, and openspec/changes/seam-verification/metrics.md (the census that golden-output-shape.test.ts:29-30 declares as its fixture source) — and no CI gate asserts census freshness. The address-feedback-diff-path-producer change refreshed only the test files (test labels 339/363/541/801 vs census 314/338/473/733), and forward-aggregate-correlation-id repeated the miss across all five edited commands, diverging every refreshed row from the census. Refresh scope: rg 'uf\.[a-z-]+\.md:[0-9]' across .opencode, openspec/, docs/, and .uf/, then reconcile every location against the command files themselves (the source of truth), not against the other label files.
