---
tag: spec-traceability
author: jay-flowers
category: gotcha
created_at: 2026-11-04T14:52:13Z
identity: spec-traceability-20261104T145213-jay-flowers
tier: draft
---

When an OpenSpec delta adds MUST prose or THEN/AND clauses mandating exact user-visible message content (error text, guidance phrasing), verify at authoring and review time that a test pins that content: grep the mandated distinctive token across the test suite — zero matches means the requirement has no enforcement path and will silently regress on refactor. Concrete case (PR 949, unbound-force/unbound-force): the build-review-prompts FR-011 delta mandated that the diff_not_found message quote the path and say to copy it exactly rather than retype it; production emitted the guidance (review-dispatch/index.ts diffPathFailure, shared by build_review_prompts and read_diff_chunk) but grep for "retype" across .opencode/test returned zero matches — build-review-prompts.test.ts asserted only the error code, and read-diff-chunk.test.ts asserted the code and path quoting but not the guidance. Corollary to meta-prompt pack MP-003: a scenario whose THEN-clause no test executes is a scenario that cannot fail.
