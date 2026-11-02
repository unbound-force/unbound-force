---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:59:51Z
identity: issue-triage-20261102T025951-jay-flowers
tier: draft
---

Triage gotcha for Constitution Principle VI (Deterministic Execution) prose→tool extraction issues: diff the issue's description of current behavior against the actual command prose before rating it VALID. In unbound-force/unbound-force#675 (create_fix_branch tool for uf.review-pr), the issue cites uf.review-pr.md lines 829-993, but that range now holds the Step 8 review-posting section — the fix-branch workflow sits at ~640-828 — and the issue misstates a load-bearing detail: it says the fix branch is created "from PR head" (tool description, procedure bullet, and the "PR head SHA not found" error example), while the command creates it from the base branch (git checkout <baseRefName> then git checkout -b) for pre-existing CI failures, per the commit template "pre-existing on <baseRefName> and unrelated to PR #<PR_NUMBER>". Because the only behavior-parity acceptance criterion covers sanitization, tests written from the issue would assert the wrong branch point and pass green — a test that masks a behavior regression. Rule: parity criteria in extraction issues MUST enumerate every preserved observable behavior (branch point, dirty-tree guard, collision check), and drifted line references are a stale-description signal warranting a content diff before VALID.
