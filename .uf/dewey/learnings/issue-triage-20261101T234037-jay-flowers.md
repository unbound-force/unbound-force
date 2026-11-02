---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:40:37Z
identity: issue-triage-20261101T234037-jay-flowers
tier: draft
---

Census-drift bug reports are triage-tractable without repo access: when a report cites specific files, line numbers, and disagreeing counts across canonical label locations (unbound-force issue #823: golden-output-shape.test.ts says 14 vs 15 internally, metrics.md 47 vs golden header 48), the claim is objectively verifiable by grepping the label pattern (e.g., uf.*.md:<line> site labels) across the census locations (golden-output-shape.test.ts, dispatch-review-runs.test.ts, metrics.md registry) and recounting against the command files as source of truth — classify VALID / bug / objective. Return a null split recommendation: the census-discipline invariant requires every label location to agree in the same commit, so splitting the correction across issues or commits reproduces the drift being reported. Cross-checking the issue's provenance against the knowledge base (the validated seam-verification compiled article independently recorded the residual count drift on its Open gaps list) converts author claims into corroborated evidence before the verdict is issued.
