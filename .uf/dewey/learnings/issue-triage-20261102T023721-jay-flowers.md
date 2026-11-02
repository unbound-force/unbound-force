---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:37:21Z
identity: issue-triage-20261102T023721-jay-flowers
tier: draft
---

Documentation-gate tracking issues stay VALID for the change they track even when a later change supersedes part of the documented behavior: unbound-force/unbound-force#717 (docs tracker for dispatch-finalize-determinism) was followed by fix-dispatch-verdict-derivation (issue #764), which fixed the top-level verdict-derivation layer the first change left intact — that impact belongs to the follow-up change's own gate disposition, not to retro-scoping or invalidating the earlier issue. The corroboration chain that validates a gate issue against repo evidence: the AGENTS.md documentation-gate rule plus a user-facing change (slash-command rewrites under .opencode/commands/, mirrored to internal/scaffold/assets/opencode/), the originating change's tasks.md Docs phase mandating the filing (dispatch-finalize-determinism task 4.4), and the issue body citing the change directory with its FR range. When tasks.md omits the filed issue number, matching creation timing (issue #717 opened the same day as the incident session named in the proposal) closes the traceability loop; a direct grep of the rewritten commands for the new pattern (consolidate_dispatch one-call/provenance/dual-mode language) confirms the user-facing surface landed. Note that glob tools that skip dot-directories will miss .opencode/ paths — use explicit-path grep or read to verify them.
