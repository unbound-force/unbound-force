---
tag: test-assertions
author: jay-flowers
category: pattern
created_at: 2026-11-05T15:04:00Z
identity: test-assertions-20261105T150400-jay-flowers
tier: draft
---

Partial-match assertions cannot pin a documented output shape. When a tool's section payload is a spec-MUST contract (RD-FR-057's summary: command, mode, verdict, verdict_reason, run_counts, coverage, workflow_result, entry counts), the section's happy-path test must enumerate every documented field — a toMatchObject over a subset lets a field-drop regression pass green, and the statements-coverage floor stays green too because the producing branch still executes. Pin the full key set: extend the matcher to every documented field, or assert Object.keys(result) equals the documented set, as golden-output-shape.test.ts does for the RD-FR-049 consolidate summary. Verified on PR 949's read-dispatch-artifact.test.ts summary test, which pinned 6 of 11 documented fields; complement to test-assertions-20261104T144150 (length-only assertions vs retention policy).
