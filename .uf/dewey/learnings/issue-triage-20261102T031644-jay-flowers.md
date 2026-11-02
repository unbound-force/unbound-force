---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:44Z
identity: issue-triage-20261102T031644-jay-flowers
tier: draft
---

Issue #668 (resolve_conflict plugin tool) triage verification: the issue's uf.finale.md citation (lines 625-953, ~146 lines) is stale against current HEAD — PR #745's prose-to-tool migration shrank the file to 399 lines total and the "7b. Conflict Recovery" five-option flow now sits at ~302-350 — but the signature content is verbatim intact (five-option question-tool gate; Option 1 git_fetch→git_merge→git_push or git_merge_abort; Option 2 git_fetch→git_rebase→git_push_force_with_lease or git_rebase_abort; Option 3 Stop; Option 4 Continue; Option 5 cobalt-crush-dev sub-agent spawn), so the extraction proposal remains VALID. Key structural observation: the uf-workflow plugin already ships the atomic git tools the prose chains (git_merge, git_merge_abort, git_rebase, git_rebase_abort, git_conflict_files, git_commit_no_edit), so the proposed resolve_conflict is a DE-005 intent-level composition of existing atoms — the strategy enum is the judgment product passed at the tool boundary per DE-006 — not new git choreography. Its DE-004 token-savings estimate (~550 tokens) must be re-measured at change time since the cited prose has already shrunk. Substantial line-citation drift with verifiably intact signature content is a plan-time correction, not a NEEDS-CLARIFICATION blocker.
