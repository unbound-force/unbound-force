---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T05:25:03Z
identity: determinism-20261115T052503-jay-flowers
tier: draft
---

Hybrid classification gates split at the tool boundary. When a fix adds a step that classifies structured inputs (uf.review-pr Step 8a: "a review is a true duplicate only when commit_id == head_sha AND the finding set has not materially changed"), the evidence-gathering sub-procedures — SHA/id equality checks, extracting the prior review's finding set from the fetched review body, computing resolved/new/severity deltas — are fixed procedures that belong in a tested tool returning structured fields; only the materiality threshold and the human gate stay in prose. Shipped precedent in the same command: resolve_review_mode already owns prior-review commit-and-finding analysis for the light-check path (Step 3.11: "The tool owns the decision; do not restate, recompute, or override it"). A full-path classification step should extend that tool rather than re-instruct the comparison in prose, and the motivating incident (issue #1055: stale same-verdict review misclassified as duplicate, gate recommendation auto-accepted at 0.44 confidence) becomes the tool's regression test. Prose-executed classification reproduces the drift class the fix targets and leaves the agent parsing prior-review bodies that a tool could parse deterministically.
