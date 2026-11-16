---
tag: instruction-quality
author: jay-flowers
category: context
created_at: 2026-11-15T16:26:43Z
identity: instruction-quality-20261115T162643-jay-flowers
tier: draft
as_of: "unbound-force@f1393b41a3438503096e2003e0f6801a157d2be7"
---

In Linguist Spec Review Mode, an OpenSpec change whose entire proposed footprint is plugin TypeScript source (even when it edits embedded tool-description prose), JSON schemas, Go/TS tests, and CHANGELOG.md proposes no new or changed instruction files — the Linguist's instruction-files table is path-based and excludes TypeScript source, so the spec-review gate resolves to a clean zero-finding APPROVE. Tool-description accuracy inside plugin code is enforced by the change's own tasks and by code review (the Architect owns extracted-tool code quality; the Determinist owns whether steps belong in tools), not by the Linguist. Verified on review-dispatch-errors-cap-loses-unrecov (fixes #1039): its only prose corrections are two tool descriptions in .opencode/plugins/review-dispatch/index.ts, and design D7 / task 3.2 explicitly exclude command workflow prose changes (DE-001).
