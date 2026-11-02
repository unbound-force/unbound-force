---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:37:01Z
identity: issue-triage-20261102T023701-jay-flowers
tier: draft
---

Documentation-tracking issues in unbound-force repos that enumerate plugin tool surfaces can be fact-checked with zero I/O against the reviewing agent's own live tool registry: uf init-deployed plugins (uf-workflow, uf-finale) register their tools into the OpenCode session, so claimed tool names and counts are directly visible — issue #710's 8 uf-finale tools and 10 uf-workflow git tools all appear in the session tool list, confirming the issue describes shipped capability rather than aspiration. Two corroborating triage signals: (1) scope-update comments from follow-up OpenSpec changes legitimately extend an open docs issue's scope — fix-tui-git-output-leak folded envelope-only output and the uf_finale_stage {fileCount, files} return into #710 via comment, and the original finale-token-reduction spec marks the superseded design points accordingly — so such comments are the issue's live scope, not duplicate drift; (2) sibling docs issues covering different tools from different changes (#739's unleash-summary tooling, #810's uf_finale_base_state) remain distinct under the per-change documentation-gate pattern even though they converge on the same /uf.finale documentation surface.
