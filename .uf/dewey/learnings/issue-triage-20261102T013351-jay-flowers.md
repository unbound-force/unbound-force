---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:33:51Z
identity: issue-triage-20261102T013351-jay-flowers
tier: draft
---

Docs-tracking issues that enumerate a plugin's new tool surface can be verified without reading code, against two sources available at triage time: the implementing PR summary indexed in Dewey and the tool registrations carried in the parented triage context. Verified on unbound-force/unbound-force #766 (review-dispatch recovery tooling): all five enumerated behaviors — discard_orphaned_submissions and redispatch_failed_runs recovery tools, consolidate_dispatch success-scoping with orphaned_submission warnings, dispatch_status per-run usage reporting, dispatch_agent_run session_metadata replace-on-supply — matched PR #776's summary (closes #751) and the live tool registrations, whose descriptions repeat the DR-FR-001/DR-FR-002 identifiers and the required-vs-optional correlation_id contract. The gap half ("no in-repo pages document plugin tools") verified against the repo's documented docs/ layout (ADRs, no plugin-tool reference). Gotcha: an implementing PR that shipped a CHANGELOG entry does not satisfy or shrink the docs-tracking issue — the CHANGELOG records the change while the issue tracks operator-facing reference documentation that does not yet exist, so the issue stays VALID with full scope. Classification boundary: absent docs for a newly shipped surface is an enhancement; a doc that misstates a tool-enforced rule is a bug.
