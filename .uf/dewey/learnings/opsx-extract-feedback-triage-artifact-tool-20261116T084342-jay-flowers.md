---
tag: opsx-extract-feedback-triage-artifact-tool
author: jay-flowers
category: pattern
created_at: 2026-11-16T08:43:42Z
identity: opsx-extract-feedback-triage-artifact-tool-20261116T084342-jay-flowers
tier: draft
---

When /uf.unleash's code-review dispatch must run in a session that previously died of a context-thrash storm, the survival pattern is keeping the orchestrator's context lean rather than shrinking the review: delegate the remaining implement task and every review-fix application to cobalt-crush-dev child sessions (prompt-only, judgment-classified, with the findings quoted verbatim), run soft-gate checks with output redirected to /tmp/opencode logs and only tails kept in context, read dispatch artifacts exclusively through read_dispatch_artifact paging (never the file-read tool, whose line limit cuts long artifact lines), and skip informational subagent phases with documented notes — the gaze-reporter Phase 1b invocation fails with "Subagent depth limit reached (1)" in a parented session, and hand-running the gaze CLI through the orchestrator's context would reproduce the storm the skip avoids. With this discipline the full Step 6 fix loop (7-reviewer dispatch, REQUEST CHANGES with 5 findings, delegated fix with red/green-verified regression tests, fresh soft gate, full re-dispatch, APPROVE WITH ADVISORIES) completed in the same session that resumed the aborted run.
