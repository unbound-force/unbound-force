---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:55:54Z
identity: issue-triage-20261102T035554-jay-flowers
tier: draft
---

Mixed replacement+addition tasks can strand their additive half when dispositioned as no-ops. In unbound-force, the uf-workflow-checkpoint change's task 3.4 ("Update uf.review-pr.md: Add checkpoint calls per step... key: pr-<N>") was closed as a no-op because the file lacked the execution-checklist pattern being replaced — a rationale covering only the replacement half, silently skipping the additive checkpoint/get_progress/TodoWrite half as well. Result: uf.review-pr.md is the only long-running UF workflow command with no progress tracking (uf.unleash.md:74, uf.review-council.md:42, uf.fix-ci.md:62, uf.address-feedback.md:60, and uf.finale.md:47 all carry TodoWrite Progress Tracking sections; uf.unleash.md frames the checkpoint tool as the durable record of truth with TodoWrite as a display aid). In-file markdown step checklists ("- [ ] Step") were deliberately removed from all commands in that migration per Constitution Principle VI / determinism DE-001. When triaging command-file issues: (1) verify both halves of mixed replacement+addition tasks actually completed, and (2) reconcile the issue's proposed remedy against the current canonical pattern before adopting it — an issue describing a superseded pattern can still be VALID on its problem statement.
