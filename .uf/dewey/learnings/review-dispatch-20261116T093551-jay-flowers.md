---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-16T09:35:51Z
identity: review-dispatch-20261116T093551-jay-flowers
tier: draft
as_of: "unbound-force@62e674575e63293fa77d6db177303ca5a4fccfde"
---

Divisor child reviewers dispatched into a workspace pinned at the PR base SHA (head not checked out) get empty evidence from their deterministic Step 1 tools: `instruction-lint --base <base> --changed` reports `delta.files: []` and `tool-inventory --seams --base <base>` reports `seams: []`, because the working tree equals the merge base. Empty delta/seams there mean "the linter cannot see the head", not "no changes". Detect the condition first (the merge base equals HEAD, and reading a known-changed file from the working tree shows base content), then fall back to the persona's graceful-degradation path: review from the caller's diff manifest chunks, assess DE-007 by reading whether the change's own tests reference each changed tool by layer, corroborate DE-004 base-side numbers against the base-state linter output, and state the scope limitation in the review summary. Do not check out the PR head from a child session — sibling reviewers may share the working tree, and the shared checkout is not yours to mutate.
