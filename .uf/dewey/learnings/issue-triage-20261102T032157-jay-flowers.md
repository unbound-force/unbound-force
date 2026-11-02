---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:21:57Z
identity: issue-triage-20261102T032157-jay-flowers
tier: draft
---

Issue triage must assess proposals against current repository state, not the creation-time snapshot: an issue can be substantively delivered by a later change that never references it, which duplicate-issue searches cannot catch. Verification pattern: (1) grep the repo for the issue's proposed identifiers — zero matches is a supersession signal, not proof; (2) read the files the issue says contain the offending prose to check whether that prose still exists; (3) search the spec/knowledge base for the same capability under different names. All three converging establishes "valid but already delivered" — the correct disposition is closing as completed/superseded with a cross-reference to the delivering change, not implementation. Verified on unbound-force issue #666 (proposed wait_for_ci/check_mergeability plugin tools to replace uf.finale.md CI-polling and mergeability prose): repo-wide grep found neither identifier anywhere; the rewritten uf.finale.md (399 lines, down from ~1,089) calls uf_finale_ci_poll with mergeability folded into its typed result; openspec/changes/finale-token-reduction/ (Phase 3 of the same epic #659) delivered that design without citing #666. Gotcha: proposed designs evolve during delivery (different plugin, folded responsibilities), so match on capability and consuming-file state rather than proposed identifiers alone.
