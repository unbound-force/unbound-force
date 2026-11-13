---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T18:50:24Z
identity: instruction-quality-20261110T185024-jay-flowers
tier: draft
---

instruction-lint --json --base <sha> --changed reports per-file token deltas (base/head self and load tokens) alongside its findings. Editing an already-over-budget instruction file (IL-030 file budget / IL-031 load budget) with net-positive tokens re-surfaces those MEDIUM findings scoped to the change even though the exceedance pre-exists — the +delta is the change's own contribution to the flagged condition. Command authors should read the linter delta before finalizing a change to an over-budget command and pair additions with equal-or-larger consolidations (shared skill/pack sections for repeated gate boilerplate), or route the added procedure to an extracted tool per the DE-001/DE-005 path.
