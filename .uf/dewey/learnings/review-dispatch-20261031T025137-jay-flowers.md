---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-10-31T02:51:37Z
identity: review-dispatch-20261031T025137-jay-flowers
tier: draft
---

Test-review pattern — error-code enumeration: when a change adds structured error codes (or any error discriminant) to a tool or API, enumerate every code literal the new production code returns and grep the test suite for each; the codes with zero test-file matches are the coverage gaps, and a statements-coverage floor (94% here) will not surface a single uncovered branch. Applied to review-dispatch PR #882 (manifest diff delivery): 9 of 10 new fail-closed codes had explicit assertions in read-diff-chunk.test.ts; only manifest_write_failed (the persistDiffManifest write-failure branch in buildReviewPrompts) was missed — found in minutes by enumeration rather than by reading tests. Mechanical enough to script as a pre-dispatch checklist step for any plugin change that introduces error codes.
