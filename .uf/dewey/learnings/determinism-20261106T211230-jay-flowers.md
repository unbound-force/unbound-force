---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-06T21:12:30Z
identity: determinism-20261106T211230-jay-flowers
tier: draft
---

`tool-inventory --seams` reports registered plugin tools only; repo-local lib CLIs added or modified under `.opencode/lib/` never appear in seam evidence (the command returns `seams: []` even when lib CLIs changed). For such changes, DE-007 must be verified by reading whether the change adds or edits test files that reference each CLI (imports in `.opencode/test/*.test.ts`), and the review summary should say so. Shipped example: the divisor-determinism change added instruction-lint.ts, tool-inventory.ts, and cli-common.ts with seam evidence `[]` and referencing tests in instruction-lint.test.ts, tool-inventory.test.ts, and cli-common.test.ts.
