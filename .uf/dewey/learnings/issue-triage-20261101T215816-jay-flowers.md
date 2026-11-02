---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:58:16Z
identity: issue-triage-20261101T215816-jay-flowers
tier: draft
---

Documentation-gate issues triage fast and objectively when they enumerate verifiable behavioral claims. Issue #880 (docs for the review-dispatch diff-manifest change, RD-FR-032) was validated without code inspection via three checks: (1) cross-check every claim against shipped contracts and spec artifacts — the 1500-line DIFF_THRESHOLD mode switch, the 0o600 <diff_path>.manifest.json artifact, read_diff_chunk's 256 KiB chunk bound, the **Files Reviewed**: attestation, and consolidate_dispatch's non-gating empty_coverage_attestation advisory all matched the openspec change (openspec/changes/review-dispatch-diff-manifest/) and the stored learning review-dispatch-20261031T003108-jay-flowers; (2) confirm the referenced spec directory exists with completed tasks, so the behavior is shipped rather than aspirational; (3) grep user-facing docs (docs/, README.md) for the feature's identifiers (read_diff_chunk, Diff Manifest, empty_coverage_attestation, DIFF_THRESHOLD) to confirm the documentation gap is real and the issue is not duplicative of existing coverage. Docs issues that name thresholds, tool names, artifact paths, and advisory identifiers give triagers a checkable evidence surface; vague "improve the docs" requests do not, and should be sent back for specifics before implementation.
