---
tag: openspec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:27:29Z
identity: openspec-governance-20261105T162729-jay-flowers
tier: draft
---

A quick-tier branch (no spec artifacts) that changes behavior mandated by live, completed-but-unarchived OpenSpec deltas creates a silent spec-vs-implementation contradiction: nothing in the quick pipeline reconciles the deltas, so archiving them later lands MUST text that contradicts the shipped code — for fail-closed controls (skill-reference verification, push-lease safety) that corpus text can mandate the weaker prior behavior a future spec-faithful implementation would restore. The remedy is a retroactive OpenSpec change (proposal.md, design.md recording the design disposition, tasks.md, MODIFIED requirement deltas) committed on the same branch, superseding the affected requirements before the older deltas archive.
