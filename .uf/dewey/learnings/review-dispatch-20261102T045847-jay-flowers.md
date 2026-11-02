---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:58:47Z
identity: review-dispatch-20261102T045847-jay-flowers
tier: draft
---

Validated recovery pattern for whole-pipeline /uf.unleash delegations that exceed the uf_delegate_command 30-minute timeout cap, proven on branch opsx/review-pr-consolidate-dispatch (2026-11-02): the child session driving Step 6 code review was killed mid-dispatch after reviewers had already run, leaving no code-review checkpoint behind. A fresh session recovered the tail mechanically without re-running any successful reviewer. First, re-anchor on durable state — get_progress with command uf.unleash plus parse_tasks on the feature dir's tasks.md — to learn which steps are complete and must never be re-run. Second, inspect the stranded dispatch with dispatch_status using its correlation_id, because run state persists on disk under .uf/artifacts/dispatch/ even when the parent session dies. Third, re-execute only the killed run (divisor-testing, via redispatch_failed_runs / DR-FR-002, which reuses the persisted prompt and archives the prior run file; replacement run daa8dd64). Fourth, finalize with exactly one consolidate_dispatch call carrying provenance (branch, 40-character head SHA, workflow_id uf.review-council), which persists the approving artifact. Finally, checkpoint the recovered code-review step from the persisted artifact. Key insight: durable checkpoint state plus on-disk dispatch artifacts make a killed long delegation recoverable end-to-end — resume from durable state, recover the dispatch through its own recovery tools rather than re-running reviewers, then record the recovered step from the artifact.
