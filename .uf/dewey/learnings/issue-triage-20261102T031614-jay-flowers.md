---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:14Z
identity: issue-triage-20261102T031614-jay-flowers
tier: draft
---

Triage rule for prose→tool extraction issues whose cited command prose has drifted: classify the drift by what the acceptance criteria assert. Load-bearing drift (sibling issue #675): criteria assert parity with the cited prose — e.g., a preserved branch point — so drift corrupts the tests an implementer writes, and the issue needs correction before VALID. Cosmetic drift (issue #669): criteria define the forward-looking tool contract (typed outputs, internal pagination, atomic operations), so drift in the problem statement is plan-time context only — the implementer replaces the current prose regardless of form. Verified on unbound-force/unbound-force #669 (ensure_labels/fetch_issue_comments tools): the cited `--paginate --slurp` + JQ flatten/map snippet no longer exists — the dispatch-finalize-determinism change replaced it with a portable `--jq` projection after a gh-CLI incompatibility, and scaffold contract test FR-006 (internal/scaffold/command_contract_test.go) now bans `--slurp` in command files — and `gh label list` is absent from the current label procedure; yet both extraction targets persist (uf.triage-issue.md §1.1 paginated comment fetch and §4.2 label-ensure choreography; uf.address-feedback.md lines 146-159 paginated fetches), so the issue remains implementable as written. Drift incidents are also affirmative extraction evidence: every gh-choreography bug fixed in prose demonstrates the fragility that tool-internal choreography eliminates.
