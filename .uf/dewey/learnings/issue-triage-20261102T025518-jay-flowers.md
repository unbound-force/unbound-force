---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:55:18Z
identity: issue-triage-20261102T025518-jay-flowers
tier: draft
---

Documentation-issue triage in plugin families must compare the specific capability being documented, not the plugin or change-family name. The unbound-force review-dispatch plugin accumulated tools across successive OpenSpec changes (fetch_pr_review_context from opsx/fetch-pr-review-context, then fetch_pr_checks and fetch_failed_run_logs from extend-fetch-pr-review-context), each with its own docs issue. Keyword searches over plugin names surface these sibling issues as duplicate candidates; they are related but not duplicates because the documented contract surface (which tool, which input/output schemas, which error codes) differs. During triage of docs issues, extract the exact tool/schema/error-code surface from the issue body and compare it against each candidate's body before classifying as duplicate. Verified on issue #696 (fetch_pr_review_context docs): candidates #814 (/uf.address-feedback diff producer) and #873 (fetch_pr_checks/fetch_failed_run_logs + /uf.fix-ci re-route) both matched the "review-dispatch" keyword space but documented different capabilities.
