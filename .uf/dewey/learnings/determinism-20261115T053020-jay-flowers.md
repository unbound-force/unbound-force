---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T05:30:20Z
identity: determinism-20261115T053020-jay-flowers
tier: draft
---

Spec-review trap for changes that modify an existing tool's output contract: tasks.md gets contract tests for the new result fields (output-shape layer) while the round-trip layer — persisted state the tool writes and a later call reads back, such as checkpoint fields read via get_progress — goes untasked. DE-007's three-layer discipline (registration surface, documented output shape, round-trips) is scoped to prose→tool migrations for the HIGH bar, but modified tool surfaces create the same seams; when reviewing tasks.md for a tool-surface change, check that all three layers are named, not just the contract tests. A round-trip gap tasked at spec stage is a MEDIUM auto-fix; the same gap shipped at code review is a CRITICAL seam failure.
