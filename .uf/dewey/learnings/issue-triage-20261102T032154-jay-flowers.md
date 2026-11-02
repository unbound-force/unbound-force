---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:21:54Z
identity: issue-triage-20261102T032154-jay-flowers
tier: draft
---

Tool-extraction issues can be obsolete even when a grep for the proposed tool names returns zero matches: sibling changes under the same parent epic may have shipped the same functionality under different names in different plugins. Verified on unbound-force/unbound-force issue #666 (proposed wait_for_ci + check_mergeability in uf-workflow, parent #659 Phase 2 — GitHub API Tools): a name-grep for both proposed tools returned nothing, yet the finale-token-reduction OpenSpec change shipped `uf_finale_ci_poll` in the uf-finale plugin — implementing the internal CI polling loop AND the mergeability pre-check (gh pr view --json mergeable,mergeStateStatus, CONFLICTING/DIRTY fast-path) — rewrote uf.finale.md from ~1,089 to ~300 lines to call the tool (the bash polling loop and 36-line mergeability prose no longer exist), and its docs issue #710 closed epic #659. Recipe extension to the existing tool-extraction triage recipe (issue-triage-20261102T025933): when the proposed-name grep is empty, (1) read the parent epic's executing OpenSpec changes and match their capability lists functionally against the proposal, not by name; (2) grep the consuming command file for tool invocations that replaced the cited prose (uf.finale.md:278 now calls uf_finale_ci_poll where the loop lived); (3) check docs/closing issues for "closes <parent-epic>" references. Nominal-duplicate checks miss functional supersession. Residual unimplemented details (e.g., UNKNOWN-retry on mergeability, which uf_finale_ci_poll swallows and skips) belong in a new narrowly-scoped issue against the shipped tool — not a VALID verdict for the obsolete proposal.
