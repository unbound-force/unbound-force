---
tag: uf-unleash
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:59:45Z
identity: uf-unleash-20261102T055945-jay-flowers
tier: draft
---

Running /uf.unleash end-to-end on init-override-stubs-gitignore-sync (branch opsx/init-override-stubs-gitignore-sync, 2026-11-02) surfaced three pipeline disciplines that kept a 25-task, 7-phase, two-issue changeset (unbound-force/unbound-force #903 + #902) on rails without human intervention. First, delegation prompts that carried the spec-review advisories forward as explicit "Reviewer advisory from the approved spec review" addenda produced code that honored the advisories on first pass — the code review later confirmed the failed-disposition error handling (A1) and CRLF normalization/comment-sentinel choices (A2) were correctly implemented, converting spec-phase findings into implementation-phase guarantees. Second, on a child DELEGATE_FAILED timeout (task 5.3), verifying the working tree and tasks.md checkbox before re-delegating revealed the child had completed all work before its final-report timeout — blindly re-delegating would have duplicated effort or clobbered a finished tree. Third, after an intervening auto-compaction left a compression summary claiming Phase 6 was unstarted, the durable-evidence re-anchor (checkboxes [x], git status showing the docs diffs, empty worktree list) proved the phase complete and the summary stale — compression summaries are navigation aids, never progress authorities. The general pattern: every mid-flight recovery question resolves by reading the durable artifact (tasks.md checkboxes, git status, checkpoint phases map, dispatch artifacts) rather than trusting any in-context narrative, including one's own summaries.
