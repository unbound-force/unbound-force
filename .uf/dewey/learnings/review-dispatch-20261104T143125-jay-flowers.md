---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:31:25Z
identity: review-dispatch-20261104T143125-jay-flowers
tier: draft
---

In the review-dispatch plugin, spec requirement bodies enumerate more fail-closed conditions than their scenario blocks repeat: RD-FR-057's body requires rejection of artifacts over 64 MiB (ARTIFACT_MAX_FILE_BYTES, plugins/review-dispatch/index.ts), but its 'Reader fails closed' scenario lists only no-artifact / unparseable / symlink / schema-failure — so the scenario-first suite added in PR #949 (.opencode/test/read-dispatch-artifact.test.ts) covered every scenario condition and still left the size branch unexercised. Aggregate coverage floors (94% statements, 85% branches in .opencode/vitest.config.ts) cannot surface a single uncovered branch. When testing a bounded reader, derive the test list from the requirement body's full condition enumeration, not from the scenarios. Size-limit branches are cheap to test because the bound is checked on lstat size before readFile: fs.truncate the artifact to bound+1 bytes (sparse) and assert the fail-closed error shape. Recurrence of learning/review-dispatch-coverage-20261009T031707-jay-flowers (build_review_prompts MAX_PROMPT_BYTES), which shows the prior lesson alone did not prevent the pattern.
