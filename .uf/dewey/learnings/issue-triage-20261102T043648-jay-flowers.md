---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:36:48Z
identity: issue-triage-20261102T043648-jay-flowers
tier: draft
---

Triage verification for gap-claim issues against unbound-force review tooling (issue #563, intent-satisfaction check): the claim verified at exact locations — uf.review-council.md Phase 1c marks Protocol 2 (Issue Linking) conditional on an explicit PR number and skipped entirely for pre-PR local councils (the /uf.unleash Step 6 path), and auto-detected PR linking is informational-only with silent skip on gh failures — but existing infrastructure is richer than the issue's framing suggests: divisor-guard already consumes linked-issue ACs when Protocol 2 fires, and issue numbers already thread from /uf.unleash intake to the PR body (route-decision checkpoint issue_number, .openspec.yaml originating_issue, uf.finale Closes #N). The actionable gap is consumption at review time plus gate semantics, not absence of threading. Dewey semantic search also surfaced in-flight PR #631 (strengthen acceptance-criteria validation in review pipeline, Closes #563) implementing rigorous/best-effort evaluation modes and the IMPLEMENTATION_DEVIATION escape hatch — knowledge-graph cross-referencing during triage reveals implementation status that issue text alone cannot.
