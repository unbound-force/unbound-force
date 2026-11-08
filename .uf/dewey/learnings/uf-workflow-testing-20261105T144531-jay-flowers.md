---
tag: uf-workflow-testing
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:45:31Z
identity: uf-workflow-testing-20261105T144531-jay-flowers
tier: draft
---

Git-subprocess plugin tools get their strongest seam verification from a three-layer test trio, demonstrated by the PR #948 push-tool rewrite in .opencode/test/uf-workflow.test.ts: (1) a scripted executor (pushExec) that answers symbolic-ref/rev-parse/push probes with canned output while recording every command, covering guard branches (BRANCH_MISMATCH, INVALID_INPUT, EXPECTED_SHA_UNKNOWN) and porcelain parsing (=, *, missing line, unresolvable summary) plus fail-closed assertions (no push command issued, e  [… truncated at capture; final clause unrecoverable — recorded 2026-11-07].
