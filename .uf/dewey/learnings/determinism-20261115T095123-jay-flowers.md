---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T09:51:23Z
identity: determinism-20261115T095123-jay-flowers
tier: draft
---

tool-inventory --seams classifies a test file's DE-007 round-trip layer by filename regex (/round-?trip/) or a test title pairing the tool name with "round trip" (classifyTest, .opencode/lib/tool-inventory.ts). Genuine write→read round-trip coverage living in behavior-named files — e.g. dispatch-review-runs.test.ts drives dispatch_review_runs, then reads the persisted run files back through consolidate_dispatch and asserts the consolidated payload — classifies as "other", so `missing: ["round-trip"]` on a seam entry can be a false negative. The inventory's header marks the layer grouping "evidence only; whether a missing layer is a finding is judgment": before reporting a DE-007 round-trip gap, read the referencing test files for actual write→read coverage, and confirm the change actually creates a new write→read seam (response-only output fields, such as dispatch_review_runs's all-failed `message`, create none).
