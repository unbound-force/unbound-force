---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:55:14Z
identity: issue-triage-20261102T025514-jay-flowers
tier: draft
---

Triage of documentation-gate issues in unbound-force/unbound-force follows a verifiable four-point pattern: (1) the named tool exists in the plugin source (.opencode/plugins/<plugin>/index.ts) and often in the uf init scaffold mirror (internal/scaffold/assets/opencode/plugins/); (2) the cited spec path exists under openspec/changes/; (3) the shipping CHANGELOG entry references the issue as "Docs: #N", proving it was filed per the AGENTS.md documentation gate; (4) the docs gap is real — grep the tool name in docs/ (user-facing reference docs live in docs/usage.md, cli-reference.md, configuration.md, architecture.md; learnings-review-*.md are internal retrospectives and do not count as coverage). Duplicate detection for plugin-family docs issues MUST match on the specific tool name, not the plugin name: #696 (fetch_pr_review_context), #873 (fetch_pr_checks/fetch_failed_run_logs + /uf.fix-ci re-route), and #814 (/uf.address-feedback §1.4 diff producer) all concern the review-dispatch plugin family but are complementary, not duplicative. Gotcha: a docs issue filed at original-ship time can describe a stale contract — #696 says pr_number is required (positive integer 1-999999), but the later extend-fetch-pr-review-context change made pr_number optional with current-branch resolution and added five PR-repair output fields. The documentation work must describe the current tool contract, so triage should note scope drift rather than treat it as invalidity.
