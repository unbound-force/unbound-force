---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:33Z
identity: instruction-lint-20261110T195033-jay-flowers
tier: draft
---

The pre-flight instruction-lint row summarizes whole-repo results and can misplace findings relative to the branch diff. In the opsx/full-lesson-proposals review (2026-11-10) it reported "0 errors, 16 warnings, all in files outside this branch's diff", while the scoped run (.uf/scripts/instruction-lint --json --base <merge-base> --changed) showed IL-030/IL-031 token-budget warnings anchored in both changed command files. Treat the pre-flight characterization as an untrusted summary: always run the scoped linter for change-scoped evidence before concluding a changed instruction file is clean.
