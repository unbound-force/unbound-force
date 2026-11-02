---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:34:43Z
identity: issue-triage-20261102T013443-jay-flowers
tier: draft
---

Triage duplicate-detection for documentation and tutorial issues must search across repos, not only the current repository's issue list. One change's documentation surface is typically tracked in parallel: the in-repo docs issue (repo documentation gate), the website-repo docs issue (cross-repo documentation gate), and follow-up tutorial issues. Triaging unbound-force/unbound-force #767 (tutorial: recovering a stranded review dispatch) via current-repo keyword search surfaced only #766 (docs: review-dispatch recovery tools and consolidation changes); a Dewey semantic search additionally surfaced unbound-force/website#313 (docs: review-dispatch recovery tools, issue #751 fix) covering near-identical workflow content — dispatch_status with explicit correlation_id from a fresh parent session, choosing discard_orphaned_submissions vs redispatch_failed_runs, re-running consolidate_dispatch with attention to orphaned-submission warnings and the superseded/ run archive. Rule: when triaging any docs/tutorial issue, query the knowledge graph or the website repo for the change's documentation-gate issues before ruling on duplicates. Distinct deliverable forms (reference docs vs operational tutorial) are valid-but-overlapping, not strict duplicates: keep open with explicit scope partition and bidirectional cross-links, honoring the zero-waste mandate and preventing divergent duplicated documentation.
