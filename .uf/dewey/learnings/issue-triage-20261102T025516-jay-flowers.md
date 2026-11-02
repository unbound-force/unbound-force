---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:55:16Z
identity: issue-triage-20261102T025516-jay-flowers
tier: draft
---

When triaging documentation-gate issues for plugin tools, follow-up "extend-" changes create two traps. (1) Duplicate discrimination: sibling docs issues from the extend change name different tools in the same plugin — e.g., unbound-force #873 (fetch_pr_checks/fetch_failed_run_logs, from extend-fetch-pr-review-context) vs #696 (fetch_pr_review_context, from fetch-pr-review-context) — so plugin-family keyword overlap is not duplication; the subject tool, not the plugin, discriminates. (2) Contract-version drift: the extend change may also have MODIFIED the original tool (extend-fetch-pr-review-context made pr_number optional, added current-branch resolution and PR-repair output fields), so a docs issue citing the original spec must be implemented against the tool's current contract, not the cited spec's original FR set. Cross-check the live tool registration surface (the dispatched reviewer's own MCP registry exposes the current input/output schema) to confirm both existence and current shape before scoping the docs work.
