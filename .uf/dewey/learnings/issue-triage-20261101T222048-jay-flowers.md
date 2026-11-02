---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:20:48Z
identity: issue-triage-20261101T222048-jay-flowers
tier: draft
---

Documentation-gate issues (label: docs) filed per the AGENTS.md documentation-gate rule triage cleanly as VALID/enhancement/objective when four checks pass: (1) the gate applies — the underlying change ships user-facing surfaces (owner-facing config schema fields, MCP tool I/O contracts, slash-command prose), so the internal-only exemption does not apply; (2) the issue enumerates the concrete surfaces with the change's FR/DR requirement ids and references the change directory plus originating issue, rather than a generic "update the docs"; (3) the referenced OpenSpec change directory exists, its spec deltas carry the cited FR ids, the claimed CHANGELOG entry is present, and the target doc (AGENTS.md) does not yet cover the enumerated surfaces; (4) duplicate searches (primary keyword plus title-keyword fallback) return no other open issue. Filing before the feature PR merges is gate-compliant timing, not prematurity. Verified on unbound-force/unbound-force#859 (derive-per-run-timeout): all checks passed — RD-FR-009 confirmed in specs/review-matrix.md, RD-FR-029/031 and the DR-FR-002 modification confirmed in the change's design and spec deltas, the CHANGELOG entry present under Unreleased, and AGENTS.md carrying no per_run_timeout/timeout_effective/redispatch coverage.
