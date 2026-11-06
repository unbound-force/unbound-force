---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:03:57Z
identity: review-dispatch-20261105T150357-jay-flowers
tier: draft
---

When a tool's registered description documents its output shape field-by-field (e.g. read_dispatch_artifact's summary: command, mode, verdict, verdict_reason, run_counts, coverage, workflow_result, counts), the tool's dedicated contract test must pin every documented field, sourcing expected values from the shared sample fixture. vitest toMatchObject uses subset semantics — fields absent from the expected object are never checked, so a regression dropping a documented field passes silently. Observed in PR 949's read-dispatch-artifact.test.ts: the summary test pinned 6 of 11 keys, and the five unpinned keys (mode, verdict_reason, run_counts, coverage, workflow_result) were exactly the fields a silent regression could drop; no other test covered that return shape.
