---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:52Z
identity: issue-triage-20261102T031652-jay-flowers
tier: draft
---

Plugin-tool extraction issues that predate a landed extraction refactor go quantitatively stale while their DE-005 case strengthens — triage must re-measure, not invalidate. Triage of unbound-force #668 (resolve_conflict, filed 2026-10-05 citing uf.finale.md lines 625-953, ~146 lines, ~550 token savings) found the finale-token-reduction change had since shrunk the command to 399 lines with conflict recovery at section 7b (lines 302-351, ~50 lines), and the residual prose is no longer raw git commands but fixed sequences of already-tested tools (git_fetch → git_merge/git_rebase → abort-or-push → ci_poll) with judgment-free branching — the exact determinism-pack DE-005 sequence-collapse residue. Two triage consequences: (1) re-measure DE-004 efficiency deltas at plan time instead of trusting the issue's numbers; (2) recognize the composition shift — the proposed intent-level tool composes existing GitExec-injected primitives, so the test plan is primarily strategy→result contract tests plus rollback scenarios (merge conflict → git_merge_abort; rebase conflict → git_rebase_abort; force-push failure → merge fallback), reusing the established mockExec: GitExec fake pattern in .opencode/test/uf-workflow.test.ts rather than authoring new git-logic tests.
