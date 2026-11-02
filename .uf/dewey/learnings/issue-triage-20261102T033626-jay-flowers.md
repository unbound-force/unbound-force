---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:36:26Z
identity: issue-triage-20261102T033626-jay-flowers
tier: draft
---

Refactor issues that pin exact line counts and line ranges go stale as the target file evolves — verify the cited construct still exists at current HEAD before validating. Issue #662 (unbound-force/unbound-force, filed 2026-10-05) proposed extracting a ~500-line inline Task-tool subagent prompt from uf.review-pr.md (cited 1,239 lines, prompt at lines 350-850) into a dedicated review-pr-worker.md agent. Verified 2026-11-01: the file is 1,075 lines with zero Task-tool references; it already loads the pre-flight and review-context skills by name and dispatches the AI review through the Divisor council fan-out (dispatch-advisor skill, plan_review_dispatch, dispatch_review_runs building prompts from divisor-* agent definitions). The proposal was superseded by landed refactoring, and implementing it would have created an orphaned agent (zero-waste violation). Triage verification checklist for architecture-refactor issues: (1) re-read the cited file and compare line count; (2) grep the file for the claimed mechanism (e.g., "Task" for Task-tool subagent prompts); (3) confirm whether the proposed artifact's role is already served by existing skills/agents/tools (glob the target directory); (4) check whether the issue's acceptance criteria remain executable. An issue whose prescribed artifact would be orphaned by subsequent refactoring should be closed as superseded, not implemented.
