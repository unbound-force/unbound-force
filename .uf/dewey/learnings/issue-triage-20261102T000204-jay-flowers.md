---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:02:04Z
identity: issue-triage-20261102T000204-jay-flowers
tier: draft
---

Triage verification pattern (from #824, unbound-force/unbound-force): an issue's named file path can be wrong while its defect claim is exactly right — resolve by content anchor, not by invalidating. #824 cited `.opencode/commands/constitution-check.md` lines 43/95 for the "four org principles" enumeration, but the command file at those lines is an examples-block fence; the enumeration sits at exactly lines 43/95 of `.opencode/agents/constitution-check.md` — the agent the command delegates to (step 6) — and in its byte-identical scaffold twin `internal/scaffold/assets/opencode/agents/constitution-check.md`. Two repo mechanisms produce this path drift: command renames (the uf-command-namespace-prefix change renamed `commands/*.md` to `uf.*.md`, so pre-rename citations go stale) and command→agent delegation (commands carry usage prose; enumerated procedures live in the agent). Recipe: when a cited path does not contain the quoted content, grep the distinctive quoted phrase repo-wide and match the cited line numbers against candidates — coinciding line anchors plus identical quoted text identify the real target; then check for drift-enforced scaffold twins before sizing the fix, since the fix must land on both copies. Verdict impact: a path inaccuracy that the issue's own line-number evidence uniquely resolves does not warrant INVALID or NEEDS-CLARIFICATION.
