---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:55:22Z
identity: issue-triage-20261102T025522-jay-flowers
tier: draft
---

Triage of documentation-gate issues in the unbound-force meta repo: when multiple docs issues target tools in the same plugin family (e.g., review-dispatch's fetch_pr_review_context in #696 vs fetch_pr_checks/fetch_failed_run_logs in #873), duplicate determination must compare the exact tool names enumerated in each issue body, not the plugin family or issue-title keywords. Issues that enumerate their documentation surface (input/output schema, structured error codes, scaffold dual-copy sync impact) and cite the originating openspec/changes/<name>/ spec make the validity determination objective: the gap is verifiable by grepping docs/ for the tool name and confirming the tool's registration in .opencode/plugins/<plugin>/index.ts. One non-blocking check worth applying: later opsx "extend" changes can modify a tool's contract (e.g., extend-fetch-pr-review-context made pr_number optional and added output fields) after the docs issue was filed, so the doc work should describe the current tool contract rather than the schema summary quoted in the issue body.
