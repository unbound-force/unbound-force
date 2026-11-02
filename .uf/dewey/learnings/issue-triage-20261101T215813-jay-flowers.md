---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:58:13Z
identity: issue-triage-20261101T215813-jay-flowers
tier: draft
---

Documentation-gate issues are triage-valid and directly actionable when the body enumerates the operator-observable surfaces of the change to document rather than only naming the feature. Issue #880 (docs for review-dispatch diff-manifest delivery) models the pattern: it enumerates artifacts produced (a manifest JSON persisted 0o600 beside the saved diff), in-flow behaviors (on-demand read_diff_chunk retrieval during review), output-contract fields (the **Files Reviewed** coverage attestation), advisory signals (empty_coverage_attestation warnings at consolidation), configuration boundaries (the shared 1500-line DIFF_THRESHOLD), and the changed failure mode (over-2 MiB over-threshold diffs now proceed via manifest delivery instead of failing prompt_size_exceeded). Each enumerated surface doubles as a grep-verifiable acceptance criterion for the docs PR: triage can confirm both the documentation gap (search the docs tree for each surface term) and later completion. Docs issues that only name the feature leave validity and completion unverifiable and are candidates for NEEDS-CLARIFICATION.
